Go repo sync
===

This utility assists in syncing changes between repositories containing [Go](https://golang.org/) programs. Go [doesn't allow for relative import paths](https://golang.org/pkg/cmd/go/internal/help/#HelpImportPath), which makes contributing changes between forks/copies of repositories under different namepsaces more complicated.

An example program, `sync_priv_pub` is in the root directory, and variations of it meant for sync between specific repositories are in the [cmd](cmd) directory. The variations will have cli flags like this:

```
$ soterium_to_soteria-dag -h
Usage of soterium_to_soteria-dag:
  -all
        Sync all repos
  -k    Keep staging area after completed
  -m string
        Commit message to use (default "soterium_to_soteria-dag - Auto code sync")
  -nodep
        Skip processing of repo dependencies
  -soterd
        Sync soterd
  -soterdash
        Sync soterdash
  -sotertools
        Sync sotertools
  -soterwallet
        Sync soterwallet
  -y    Skip confirmation with user before git commit & push of synced repo contents
```

# Use

1. Build and install

    ```bash
    go install -v ./cmd/... && echo "install ok"
    ```

2. Sync repositories

    This will sync from `github.com/soterium/soterd` to `github.com/soteria-dag/soterd`
    ```bash
    soterium_to_soteria-dag -soterd -m "Fixed typo in blockdag.go"
    ```

    This will sync from `github.com/soteria-dag/soterdash` to `github.com/soterium/soterdash`, but **skip** sync of
    soterdash dependency `soterd` (`-nodep` flag).
    ```bash
    soteria-dag_to_soterium -soterdash -nodep -m "Backport census worker fix"
    ```

# Testing repo sync

1. Update the `sync_priv_pub.go` file with the repositories you want to sync

2. Build and run the command

    ```bash
    go build . && echo "build ok"
    ./sync_priv_pub -all
    ```

# What the tool does

* Process dependencies before the current repo pair (`soterd` processed before `soterwallet`)
* Clones Source and Dest repos to a staging area, and keeps it separate from your existing workspaces (`go` commands use staging area `GOPATH`)
* Syncs changes from Source to Dest using `git archive`
* Removes files from Dest that no longer exist in Source (`git rm`)
* Replaces references to Source repo tree (ex: `github.com/soterium/soterd/exp0 -> github.com/colakong/soterd/master`)
* Replaces references to Source repo path (ex: `github.com/soterium/soterd -> github.com/colakong/soterd`)
    * This handles go `import` statements
* Replaces references to dependency Source repos with dependency Dest repos
* Updates go module name to match Dest
* Replaces go module dependencies
    * Because dependencies were processed first, their new [pseudo version](https://golang.org/cmd/go/#hdr-Pseudo_versions) can be used here.
* Adds new untracked files in Dest.
* Commits changes to local Dest clone
* Pushes changes to Dest git tree (branch)