## defradb client document create

Create a new document.

### Synopsis

Create a new document.

Example: create document
  defradb client document create --collection User '{ "name": "Bob" }'

Example: create documents
  defradb client document create --collection User '[{ "name": "Alice" }, { "name": "Bob" }]'
		

```
defradb client document create --collection <collection> <document> [flags]
```

### Options

```
  -c, --collection string   Collection name
  -h, --help                help for create
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

* [defradb client document](defradb_client_document.md)	 - Create, read, update, and delete documents.

