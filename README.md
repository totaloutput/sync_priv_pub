sync_priv_pub
===

This utility assists in syncing changes between private and public repositories.

```
$ ./sync_priv_pub
Usage of sync_priv_pub:
  -k    Keep staging area after completed
  -y    Skip confirmation with user before git commit & push of synced repo contents
```

# Use

1. Update the `sync_priv_pub.go` file with the repositories you want to sync

2. Build and run the command

    ```bash
    go build . && echo "build ok"
    ./sync_priv_pub
    ```