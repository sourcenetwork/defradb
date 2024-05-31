## defradb client collection

Interact with a collection.

### Synopsis

Create, read, update, and delete documents within a collection.

### Options

```
      --get-inactive      Get inactive collections as well as active
  -h, --help              help for collection
  -i, --identity string   Hex formatted private key used to authenticate with ACP
      --name string       Collection name
      --schema string     Collection schema Root
      --tx uint           Transaction ID
      --version string    Collection version ID
```

### Options inherited from parent commands

```
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --no-log-color               Disable colored log output
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client collection create](defradb_client_collection_create.md)	 - Create a new document.
* [defradb client collection delete](defradb_client_collection_delete.md)	 - Delete documents by docID or filter.
* [defradb client collection describe](defradb_client_collection_describe.md)	 - View collection description.
* [defradb client collection docIDs](defradb_client_collection_docIDs.md)	 - List all document IDs (docIDs).
* [defradb client collection get](defradb_client_collection_get.md)	 - View document fields.
* [defradb client collection patch](defradb_client_collection_patch.md)	 - Patch existing collection descriptions
* [defradb client collection update](defradb_client_collection_update.md)	 - Update documents by docID or filter.

