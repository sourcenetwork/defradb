## defradb client collection

Interact with a collection.

### Synopsis

Add, describe, patch, set-active, delete, and truncate collections.

### Options

```
  -h, --help              help for collection
  -i, --identity string   Hex formatted private key used to authenticate with ACP
      --tx uint           Transaction ID
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client collection add](defradb_client_collection_add.md)	 - Add new collection
* [defradb client collection delete](defradb_client_collection_delete.md)	 - Delete collections
* [defradb client collection describe](defradb_client_collection_describe.md)	 - View collection version.
* [defradb client collection patch](defradb_client_collection_patch.md)	 - Patch existing collection versions
* [defradb client collection set-active](defradb_client_collection_set-active.md)	 - Set the active collection version
* [defradb client collection truncate](defradb_client_collection_truncate.md)	 - Truncate the given collection

