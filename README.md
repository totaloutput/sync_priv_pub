Go repo sync
===

This utility assists in syncing changes between repositories containing [Go](https://golang.org/) programs. Go [doesn't allow for relative import paths](https://golang.org/pkg/cmd/go/internal/help/#HelpImportPath), which makes contributing changes between forks/copies of repositories under different namepsaces more complicated.

An example program, `example.go` is in the root directory, and variations of it meant for sync between specific repositories are in the [cmd](cmd) directory. The variations will have cli flags like this:

```
$ soterium_to_soteria-dag -h
Usage of soterium_to_soteria-dag:
  -all
        Sync all repos
  -e string
        Email address to use for commit
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
  -u string
        User name to use for commit
  -y    Skip confirmation with user before git commit & push of synced repo contents
```

# Use

1. Make sure `go mod` tooling can access any _private repos_

    If one or more of the repos you are syncing is _private_, you may want to redirect git requests using https to use ssh instead. This allows `go mod`and related package-management tools to work in a situation where they can't prompt for user input (github.com user/pass). In the example below, we are redirecting requests for the private `github.com/soterium` repositories.
    ```bash
    # Add this section to your ~/.gitconfig file

    [url "ssh://git@github.com/soterium/"]
        insteadOf = https://github.com/soterium/
    ```

2. Build and install

    ```bash
    go install -v ./cmd/... && echo "install ok"
    ```

3. Sync repositories

    This will sync from `github.com/soterium/soterd` to `github.com/soteria-dag/soterd`
    ```bash
    soterium_to_soteria-dag -soterd -e banana@fogscape.net -u "Banana Man" -m "Fixed typo in blockdag.go"
    ```

    This will sync from `github.com/soteria-dag/soterdash` to `github.com/soterium/soterdash`, but **skip** sync of
    soterdash dependency `soterd` (`-nodep` flag).
    ```bash
    soteria-dag_to_soterium -soterdash -nodep -e banana@fogscape.net -u "Banana Man" -m "Backport census worker fix"
    ```

# Testing repo sync

1. Update the `example.go` file with the repositories you want to sync

2. Build and run the command

    ```bash
    go build . && echo "build ok"
    ./example -all
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
    * If an email address was specified, this is used for the commit instead of your global default.
* Pushes changes to Dest git tree (branch)
