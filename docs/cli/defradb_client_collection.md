## defradb client collection

Interact with a collection.

### Synopsis

Create, read, update, and delete documents within a collection.

### Options

```
  -h, --help             help for collection
      --name string      Collection name
      --schema string    Collection schema Root
      --tx uint          Transaction ID
      --version string   Collection version ID
```

### Options inherited from parent commands

```
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --rootdir string       Directory for data and configuration to use (default: $HOME/.defradb)
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client collection create](defradb_client_collection_create.md)	 - Create a new document.
* [defradb client collection delete](defradb_client_collection_delete.md)	 - Delete documents by docID or filter.
* [defradb client collection describe](defradb_client_collection_describe.md)	 - View collection description.
* [defradb client collection docIDs](defradb_client_collection_docIDs.md)	 - List all document IDs (docIDs).
* [defradb client collection get](defradb_client_collection_get.md)	 - View document fields.
* [defradb client collection update](defradb_client_collection_update.md)	 - Update documents by docID or filter.

