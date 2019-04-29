package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/soterium/sync_priv_pub/tools"
)

const (
	// Mode to use for directories created during Sync
	tempPathMode = 0755

	defaultGitRemote = "origin"
)

var (
	// Exclude files/directories under these paths from recursive diff comparisons between Source and Dest repos
	diffExclude = []string{".git"}

	// Exclude files under or matching these paths from renaming operations.
	// The go.mod and go.sum files will be handled with specific commands, if present.
	renameExclude = []string{".git", "go.mod", "go.sum", "glide.yaml", "glide.lock"}

	// Track which pairs have been synced already, so that sync process isn't repeated during the same run,
	// when multiple pairs share the same dependencies.
	done = make(map[string]bool)
)

type RepoPair struct {
	// The source git repository
	Source        GitRepo
	// The git tree-ish to use as the sync source, in the source git repo
	SourceGitTree string
	// The destination git repository
	Dest          GitRepo
	// The git tree-ish to use as the sync dest, in the dest git repo
	DestGitTree string
	// Pairs that should be synced before this one, due to module dependencies
	// between the involved repos.
	Dependencies  []*RepoPair
}

// Return a string representing the RepoPair
func (r *RepoPair) String() string {
	return fmt.Sprintf("%s -> %s", r.Source.String(), r.Dest.String())
}

// Sync syncs changes from the Source repo to the Dest repo,
// while also replacing references of the source in the destination.
func (r *RepoPair) Sync(keepStaging, skipAsk bool, commitMsg string) error {
	_, exists := done[r.String()]
	if exists {
		// Don't process the same repo more than once in the same run
		fmt.Printf("Skipping %s, already attempted\n", r.String())
		return nil
	}

	// Process dependencies first
	for _, dep := range r.Dependencies {
		fmt.Println("Processing dependency", dep.String())
		err := dep.Sync(keepStaging, skipAsk, commitMsg)
		if err != nil {
			return fmt.Errorf("Failed to sync dependency %s: %s", dep, err)
		}
	}

	// Create staging area. This will be used as our GOPATH dir, so we'll create a src dir inside of it too.
	staging, err := ioutil.TempDir("", "sync_priv_pub-")
	if err != nil {
		return fmt.Errorf("Failed to create staging staging: %s", err)
	}

	// If we encounter an error, we'll leave the staging area behind
	cleanup := true
	defer func() {
		// This repo was attempted to be synced
		done[r.String()] = true

		if cleanup && keepStaging {
			_ = os.RemoveAll(staging)
		}
	}()

	// Set the staging area's permissions, so that "go get" can create a
	// pkg dir inside of it when adding new dependencies.
	err = os.Chmod(staging, tempPathMode)

	fmt.Println("Created staging area at", staging)

	// Create <staging area>/src directory, which is where we will clone repositories
	srcDir := filepath.Join(staging, "src")
	err = os.Mkdir(srcDir, tempPathMode)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to create %s: %s", srcDir, err)
	}

	// Source and Dest repos will be cloned under <staging area>/src, so we'll tell go commands to search under
	// here for modules.
	goEnv := append(os.Environ(), fmt.Sprintf("GOPATH=%s", staging))

	// Clone the Source repo to the staging area
	src := filepath.Join(srcDir, r.Source.Path)
	err = r.Source.Clone(src, false)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to clone %s: %s", r.Source, err)
	}

	fmt.Println("Cloned source", r.Source.Path, "to", src)

	// Clone the Dest repo to the staging area
	dst := filepath.Join(srcDir, r.Dest.Path)
	err = r.Dest.Clone(dst, false)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to clone %s: %s", r.Dest, err)
	}

	fmt.Println("Cloned dest", r.Dest.Path, "to", dst)

	// Switch to the Dest tree, so that when we Archive/Commit to Dest, the content is being committed
	// to the correct branch regardless of what the default branch is set to.
	err = tools.GitCheckout(dst, r.DestGitTree)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to checkout to %s on %s: %s", r.DestGitTree, r.Dest.Path, err)
	}

	fmt.Println("Checked out to", r.DestGitTree)

	// Archive files from Source to Dest
	err = tools.GitArchive(src, r.SourceGitTree, dst)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to archive from %s to %s: %s", src, r.SourceGitTree, dst, err)
	}

	fmt.Println("Archived files from", src, "tree", r.SourceGitTree, "to", dst)

	// Remove files in Dest that don't exist in Source
	pruned, err := tools.GitPrune(dst, src)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to prune from %s compared to %s: %s", dst, src, err)
	}
	for _, p := range pruned {
		fmt.Printf("%s\tpruned %s\n", r.Dest.Path, p)
	}

	// Confirm that files in Source and Dest are now identical, skipping the .git directory
	same, err := tools.DiffR(dst, src, diffExclude...)
	if !same {
		cleanup = false
		return fmt.Errorf("%s != %s\n%s", src, dst, err)
	}

	fmt.Printf("Staged %s identical to %s tree %s\n", r.Dest.Path, r.Source.Path, r.SourceGitTree)

	// Replace references to source repo git tree with dest repo & git tree in dest repo, except for files under .git
	gitTreeOld := path.Join(r.Source.Path, r.SourceGitTree)
	gitTreeNew := path.Join(r.Dest.Path, r.DestGitTree)
	count, err := tools.ReplaceR(dst, gitTreeOld, gitTreeNew, renameExclude...)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to replace %s with %s in %s: %s", gitTreeOld, gitTreeNew, dst, err)
	}

	fmt.Printf("%s\tmade %d replacements for %s => %s\n", r.Dest.Path, count, gitTreeOld, gitTreeNew)

	// Replace references to source repo path with dest repo path in dest repo, except for all files under .git and go.mod files
	count, err = tools.ReplaceR(dst, r.Source.RepoPath(), r.Dest.RepoPath(), renameExclude...)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to replace %s with %s in %s: %s", r.Source.RepoPath(), r.Dest.RepoPath(), dst, err)
	}

	fmt.Printf("%s\tmade %d replacements for %s => %s\n", r.Dest.Path, count, r.Source.RepoPath(), r.Dest.RepoPath())

	// Replace references to dependency source repos with dependency dest repos, except for .git and go.mod files
	for _, dep := range r.Dependencies {
		count, err = tools.ReplaceR(dst, dep.Source.RepoPath(), dep.Dest.RepoPath(), renameExclude...)
		if err != nil {
			cleanup = false
			return fmt.Errorf("Failed to replace %s with %s in %s: %s", dep.Source.RepoPath(), dep.Dest.RepoPath(), dst, err)
		}

		fmt.Printf("%s\tmade %d replacements for %s => %s\n", r.Dest.Path, count, dep.Source.RepoPath(), dep.Dest.RepoPath())
	}

	// Determine if go module file exists in Dest repo
	goMod := filepath.Join(dst, "go.mod")
	goModExists := false
	if s, err := os.Stat(goMod); err == nil && !s.IsDir() {
		goModExists = true
	}

	if goModExists {
		// Update the go module name for the dest repo
		err := tools.GoSetMod(goMod, r.Dest.Path, goEnv)
		if err != nil {
			cleanup = false
			return fmt.Errorf("Failed to set go module name to %s in %s: %s", r.Dest.Path, goMod, err)
		}

		fmt.Printf("%s\tset module name\n", r.Dest.Path)
	}

	if goModExists && len(r.Dependencies) > 0 {
		// Update go module dependencies for dest repo

		// First, drop all old dependency names. We drop all of the old dependencies first, because "go get" attempts to
		// resolve ~all~ modules before performing its operation.
		for _, dep := range r.Dependencies {
			err := tools.GoDropMod(goMod, dep.Source.Path, goEnv)
			if err != nil {
				cleanup = false
				return fmt.Errorf("Failed to drop old go module dependency %s in %s: %s", dep.Source.Path, goMod, err)
			}

			fmt.Printf("%s\tdropped old dependency %s\n", r.Dest.Path, dep.Source.Path)
		}

		// Next, add the new dependency names
		for _, dep := range r.Dependencies {
			err := tools.GoGetMod(dst, dep.Dest.Path, goEnv)
			if err != nil {
				cleanup = false
				return fmt.Errorf("Failed to get go module dependency %s: %s", dep.Dest.Path, err)
			}

			fmt.Printf("%s\tadded new dependency %s\n", r.Dest.Path, dep.Dest.Path)
		}

		// Finally, we remove stale references to old dependencies and their related modules
		err := tools.GoTidyMod(dst, goEnv)
		if err != nil {
			cleanup = false
			return fmt.Errorf("Failed to tidy go module dependencies: %s", err)
		}

		fmt.Printf("%s\ttidied go module dependencies\n", r.Dest.Path)
	}

	if !skipAsk {
		// Ask the user if they want to commit
		fmt.Printf("About to commit changes for %s\n", r.Dest.Path)
		ok, err := tools.AskUser()
		if err != nil {
			cleanup = false
			return fmt.Errorf("Failure while asking if commit is ok: %s", err)
		}

		if !ok {
			cleanup = false
			return fmt.Errorf("User aborted")
		}
	}

	// Add files in Dest repo that weren't tracked before
	_, err = tools.GitAddNew(dst)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to git-add new files to %s: %s", dst, err)
	}

	// Commit changes
	err = tools.GitCommit(dst, commitMsg)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to commit git changes to %s: %s", dst, err)
	}

	fmt.Printf("%s\tchanges committed\n", r.Dest.Path)

	if !skipAsk {
		// Ask the user if they want to push
		fmt.Printf("About to push changes for %s to %s\n", r.Dest.Path, r.DestGitTree)
		ok, err := tools.AskUser()
		if err != nil {
			cleanup = false
			return fmt.Errorf("Failure while asking if commit is ok: %s", err)
		}

		if !ok {
			cleanup = false
			return fmt.Errorf("User aborted")
		}
	}

	// Push changes
	err = tools.GitPush(dst, defaultGitRemote, r.DestGitTree)
	if err != nil {
		cleanup = false
		return fmt.Errorf("Failed to git-push changes to %s %s: %s", r.Dest.Path, r.DestGitTree, err)
	}

	fmt.Printf("%s\tchanges pushed to %s %s\n", r.Dest.Path, defaultGitRemote, r.DestGitTree)

	return nil
}
