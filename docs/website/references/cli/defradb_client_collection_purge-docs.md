## defradb client collection purge-docs

Permanently remove documents by DocID from the local node

### Synopsis

Permanently remove a set of documents by DocID from the local node, including all
datastore values and headstore entries. When --prune-history is set, it also removes
reachable blockstore blocks that are no longer owned by another document. Shared blocks
are retained until their final owning document is purged.

History pruning is not supported for branchable collections.

Unlike the soft-delete performed by the delete command, this operation is irreversible and
does not propagate to other nodes in the peer network.

```
defradb client collection purge-docs [flags]
```

### Options

```
      --collection-id string     Collection ID
      --collection-name string   Collection name
      --docID stringArray        DocID of a document to purge (may be repeated)
      --get-inactive             Get inactive collections as well as active
  -h, --help                     help for purge-docs
      --prune-history            Also delete reachable blockstore blocks after their final owner is purged
      --version-id string        Collection version ID
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

