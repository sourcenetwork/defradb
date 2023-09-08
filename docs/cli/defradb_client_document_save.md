## defradb client document save

Create or update a document.

### Synopsis

Create or update a document.
		
Example:
  defradb client document save --collection User --key bae123 '{ "name": "Bob" }'
		

```
defradb client document save --collection <collection> --key <docKey> <document> [flags]
```

### Options

```
  -c, --collection string   Collection name
  -h, --help                help for save
      --key string          Document key
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

