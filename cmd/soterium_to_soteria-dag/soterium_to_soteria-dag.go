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
	soteriaSoterd = repo.GitRepo{Path: "github.com/soteria-dag/soterd"}

	soteriumSoterDash = repo.GitRepo{Path: "github.com/soterium/soterdash"}
	soteriaSoterDash = repo.GitRepo{Path: "github.com/soteria-dag/soterdash"}

	soteriumSoterWallet = repo.GitRepo{Path: "github.com/soterium/soterwallet"}
	soteriaSoterWallet = repo.GitRepo{Path: "github.com/soteria-dag/soterwallet"}

	soteriumSoterTools = repo.GitRepo{Path: "github.com/soterium/sotertools"}
	soteriaSoterTools = repo.GitRepo{Path: "github.com/soteria-dag/sotertools"}

	// Define the sync direction of repositories
	soterd = repo.RepoPair{
		Source:        soteriumSoterd,
		SourceGitTree: "exp0",
		Dest: soteriaSoterd,
		DestGitTree: "master",
	}

	soterdash = repo.RepoPair{
		Source: soteriumSoterDash,
		SourceGitTree: "master",
		Dest: soteriaSoterDash,
		DestGitTree: "master",
		Dependencies: []*repo.RepoPair{&soterd},
	}

	soterwallet = repo.RepoPair{
		Source: soteriumSoterWallet,
		SourceGitTree: "exp0",
		Dest: soteriaSoterWallet,
		DestGitTree: "master",
	}

	sotertools = repo.RepoPair{
		Source: soteriumSoterTools,
		SourceGitTree: "master",
		Dest: soteriaSoterTools,
		DestGitTree: "master",
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
	var keepStaging, skipAsk bool
	var syncAll, syncSoterd, syncSoterDash, syncSoterWallet, syncSoterTools bool
	flag.BoolVar(&keepStaging, "k", false,"Keep staging area after completed")
	flag.StringVar(&commitMsg, "m", defaultCommitMsg, "Commit message to use")
	flag.BoolVar(&skipAsk, "y", false,"Skip confirmation with user before git commit & push of synced repo contents")
	flag.BoolVar(&syncAll, "all", false, "Sync all repos")
	flag.BoolVar(&syncSoterd, "soterd", false, "Sync soterd")
	flag.BoolVar(&syncSoterDash, "soterdash", false, "Sync soterdash")
	flag.BoolVar(&syncSoterWallet, "soterwallet", false, "Sync soterwallet")
	flag.BoolVar(&syncSoterTools, "sotertools", false, "Sync sotertools")
	flag.Parse()

	if !syncAll && !syncSoterd && !syncSoterDash && !syncSoterWallet && !syncSoterTools {
		abort(fmt.Sprintf("Need to specify one or more repos to sync! (-soterd, -soterdash, -soterwallet, -sotertools, or -all for all repos)"))
	}

	if syncAll {
		syncSoterd = true
		syncSoterDash = true
		syncSoterWallet = true
		syncSoterTools = true
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
		err := soterd.Sync(keepStaging, skipAsk, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", soterd.String(), err))
		}

		fmt.Println()
	}

	if syncSoterDash {
		fmt.Println("Syncing", soterdash.String())
		err := soterdash.Sync(keepStaging, skipAsk, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", soterdash.String(), err))
		}

		fmt.Println()
	}

	if syncSoterWallet {
		fmt.Println("Syncing", soterwallet.String())
		err := soterwallet.Sync(keepStaging, skipAsk, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", soterwallet.String(), err))
		}

		fmt.Println()
	}

	if syncSoterTools {
		fmt.Println("Syncing", sotertools.String())
		err := sotertools.Sync(keepStaging, skipAsk, commitMsg)
		if err != nil {
			abort(fmt.Sprintf("Failed to sync %s:\n%s", sotertools.String(), err))
		}

		fmt.Println()
	}
}