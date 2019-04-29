package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/soterium/sync_priv_pub/repo"
	"github.com/soterium/sync_priv_pub/tools"
)

var (
	neededCmds = []string{"git", "tar", "diff"}

	// Repositories involved in sync process
	soteriumSoterd = repo.GitRepo{Path: "github.com/soterium/soterd"}
	//soteriaSoterd = repo.GitRepo{Path: "github.com/soteria-dag/soterd"}
	colakongSoterd = repo.GitRepo{Path: "github.com/colakong/soterd"}

	soteriumSoterDash = repo.GitRepo{Path: "github.com/soterium/soterdash"}
	//soteriaSoterDash = repo.GitRepo{Path: "github.com/soteria-dag/soterdash"}
	colakongSoterDash = repo.GitRepo{Path: "github.com/colakong/soterdash"}

	// Define the sync direction of repositories
	soterd = repo.RepoPair{
		Source:        soteriumSoterd,
		SourceGitTree: "exp0",
		//Dest: soteriaSoterd,
		Dest: colakongSoterd,
		DestGitTree: "master",
	}

	soterdash = repo.RepoPair{
		Source: soteriumSoterDash,
		SourceGitTree: "master",
		//Dest: soteriaSoterDash,
		Dest: colakongSoterDash,
		DestGitTree: "master",
		Dependencies: []*repo.RepoPair{&soterd},
	}
)

// abort prints the message and exits with code 1
func abort(msg string) {
	fmt.Println(msg)
	syscall.Exit(1)
}

func main() {
	defaultCommitMsg := fmt.Sprintf("%s - Auto code sync", tools.ThisFile())

	var commitMsg string
	var keepStaging, skipAsk, skipDeps bool
	var syncAll, syncSoterd, syncSoterDash bool
	flag.BoolVar(&keepStaging, "k", false,"Keep staging area after completed")
	flag.StringVar(&commitMsg, "m", defaultCommitMsg, "Commit message to use")
	flag.BoolVar(&skipAsk, "y", false,"Skip confirmation with user before git commit & push of synced repo contents")
	flag.BoolVar(&skipDeps, "nodep", false, "Skip processing of repo dependencies")
	flag.BoolVar(&syncAll, "all", false, "Sync all repos")
	flag.BoolVar(&syncSoterd, "soterd", false, "Sync soterd")
	flag.BoolVar(&syncSoterDash, "soterdash", false, "Sync soterdash")
	flag.Parse()

	if !syncAll && !syncSoterd && !syncSoterDash {
		abort(fmt.Sprintf("Need to specify one or more repos to sync! (-soterd, -soterdash, or -all for all repos)"))
	}

	if syncAll {
		syncSoterd = true
		syncSoterDash = true
	}

	// Look for needed commands
	for _, cmd := range neededCmds {
		_, exists := tools.Which(cmd)
		if !exists {
			abort(fmt.Sprintf("Missing needed command %s", cmd))
		}
	}

	// Sync repositories
	if syncSoterd {
		fmt.Println("Syncing", soterd.String())
		err := soterd.Sync(keepStaging, skipAsk, skipDeps, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", soterd.String(), err))
		}

		fmt.Println()
	}

	if syncSoterDash {
		fmt.Println("Syncing", soterdash.String())
		err := soterdash.Sync(keepStaging, skipAsk, skipDeps, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", soterdash.String(), err))
		}

		fmt.Println()
	}
}
