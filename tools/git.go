package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitAddNew stages new files for commit, and returns the names of those that were added
func GitAddNew(path string) ([]string, error) {
	added := make([]string, 0)
	git, exists := Which("git")
	if !exists {
		return added, fmt.Errorf("Couldn't find git command")
	}

	untracked, err := GitUntracked(path)
	if err != nil {
		return added, err
	}

	for _, n := range untracked {
		cmd := exec.Command(git, "add", n)
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return added, fmt.Errorf("%s\n%s", output, err)
		}

		added = append(added, n)
	}

	return added, nil
}

// GitArchive takes a copy of files from src tree (branch, commit, etc) and outputs them to dst
func GitArchive(src, tree, dst string) error {
	git, exists := Which("git")
	if !exists {
		return fmt.Errorf("Couldn't find git command")
	}
	tar, exists := Which("tar")
	if !exists {
		return fmt.Errorf("Couldn't find tar command")
	}

	// Issue git archive on the src, and pipe the output to tar at the dst
	archive := exec.Command(git, "archive", "--format=tar", tree)
	archive.Dir = src

	out, err := archive.StdoutPipe()
	if err != nil {
		return err
	}

	untar := exec.Command(tar, "-C", dst, "-xvf", "-")
	untar.Stdin = out

	err = archive.Start()
	if err != nil {
		return err
	}
	err = untar.Start()
	if err != nil {
		return err
	}

	err = archive.Wait()
	if err != nil {
		return err
	}
	err = untar.Wait()
	if err != nil {
		return err
	}

	return nil
}

// GitCheckout switches the checkout to the given tree
func GitCheckout(path, tree string) error {
	git, exists := Which("git")
	if !exists {
		return fmt.Errorf("Couldn't find git command")
	}

	cmd := exec.Command(git, "checkout", tree)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}

// GitCommit commits all tracked files to the repository
func GitCommit(path, msg string) (error) {
	git, exists := Which("git")
	if !exists {
		return fmt.Errorf("Couldn't find git command")
	}

	cmd := exec.Command(git, "commit", "-a", "-m", msg)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			fmt.Print(string(output))
			return nil
		}

		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}

// GitPrune issues "git rm" in path for items in path that aren't in cmp,
// and returns a list of removed items.
func GitPrune(path, cmp string) ([]string, error) {
	pruned := make([]string, 0)
	// Files in these paths are excluded from pruning.
	//
	// If you want to add items here, make sure they are the full relative path
	// to the excluded directory from the base of the repo, and not a fragment of one.
	// Good: my/secret/area
	// Bad: secret
	exclude := []string{".git"}

	git, exists := Which("git")
	if !exists {
		return pruned, fmt.Errorf("Couldn't find git command")
	}

	// This function compares files and removes ones that shouldn't exist in path
	pruner := func(n string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(path, n)
		if err != nil {
			return fmt.Errorf("Can't determine relative path for %s: %s", n, err)
		}

		// Skip pruning files that are under an excluded directory
		for _, x := range exclude {
			if IsUnder(rel, x) {
				// The file is under the excluded path, so skip it
				//fmt.Println("Skipping", rel)
				return nil
			}
		}

		o := filepath.Join(cmp, rel)
		oInfo, err := os.Stat(o)
		if os.IsNotExist(err) {
			// Remove the file, because it doesn't exist in cmp
			cmd := exec.Command(git, "rm", rel)
			cmd.Dir = path
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s\n%s", string(output), err)
			}

			pruned = append(pruned, rel)
		} else if oInfo.IsDir() != info.IsDir() {
			// Fail, because the path in one tree is a file, and in the other it's a directory
			if info.IsDir() {
				return fmt.Errorf("%s is a directory but %s is not", n, o)
			} else {
				return fmt.Errorf("%s not a directory but %s is", n, o)
			}
		}

		return nil
	}

	err := filepath.Walk(path, pruner)
	if err != nil {
		return pruned, err
	}

	return pruned, nil
}

// GitPush pushes commits to the default remote and tree
func GitPush(path, remote, tree string) error {
	git, exists := Which("git")
	if !exists {
		return fmt.Errorf("Couldn't find git command")
	}

	cmd := exec.Command(git, "push", remote, tree)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	} else {
		fmt.Print(string(output))
	}

	return nil
}

// GitUntracked returns a list of untracked files in the git repository at path
func GitUntracked(path string) ([]string, error) {
	untracked := make([]string, 0)
	git, exists := Which("git")
	if !exists {
		return untracked, fmt.Errorf("Couldn't find git command")
	}

	cmd := exec.Command(git, "status", "--porcelain=v2")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return untracked, fmt.Errorf("%s\n%s", output, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		if parts[0] == "?" {
			// This is an untracked file
			untracked = append(untracked, parts[1])
		}
	}

	return untracked, nil
}