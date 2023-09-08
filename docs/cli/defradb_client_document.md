## defradb client document

Create, read, update, and delete documents.

### Synopsis

Create, read, update, and delete documents.

### Options

```
  -h, --help   help for document
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
      --tx uint              Transaction ID
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client document create](defradb_client_document_create.md)	 - Create a new document.
* [defradb client document delete](defradb_client_document_delete.md)	 - Delete documents by key or filter.
* [defradb client document get](defradb_client_document_get.md)	 - View detailed document info.
* [defradb client document keys](defradb_client_document_keys.md)	 - List all collection document keys.
* [defradb client document save](defradb_client_document_save.md)	 - Create or update a document.
* [defradb client document update](defradb_client_document_update.md)	 - Update documents by key or filter.

