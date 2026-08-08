## defradb client document purge

Permanently remove documents by DocID from the local node

### Synopsis

Permanently remove a set of documents by DocID from the local node, including all
datastore values and headstore entries. When --prune-history is set, it also removes
reachable blockstore blocks that are no longer owned by another document. Shared blocks
are retained until their final owning document is purged.

History pruning is not supported for branchable collections.

Without --tx, logical cleanup commits one document at a time and can be resumed by
retrying. Each document must fit in one backend transaction. With --tx, the entire purge
must fit in the transaction.

Unlike the soft-delete performed by the delete command, this operation is irreversible and
does not propagate to other nodes in the peer network. It requires the node-level
purge-document permission and does not require collection or document read access.

```
defradb client document purge [flags]
```

### Examples

```
purge a document from the local node:  
  defradb client document purge --collection-name Users --docID bae-123

purge documents and their unshared history:  
  defradb client document purge --collection-name Users --docID bae-123 --docID bae-456 --prune-history
```

### Options

```
      --collection-name string   Collection name
      --docID stringArray        DocID of a document to purge (may be repeated)
  -h, --help                     help for purge
      --prune-history            Also delete reachable blockstore blocks after their final owner is purged
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

* [defradb client document](defradb_client_document.md)	 - Interact with documents.

