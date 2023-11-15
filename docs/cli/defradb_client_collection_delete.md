## defradb client collection delete

Delete documents by key or filter.

### Synopsis

Delete documents by key or filter and lists the number of documents deleted.
		
Example: delete by key(s)
  defradb client collection delete --name User --key bae-123,bae-456

Example: delete by filter
  defradb client collection delete --name User --filter '{ "_gte": { "points": 100 } }'
		

```
defradb client collection delete [--filter <filter> --key <key>] [flags]
```

### Options

```
      --filter string   Document filter
  -h, --help            help for delete
      --key strings     Document key
```

### Options inherited from parent commands

```
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --name string          Collection name
      --rootdir string       Directory for data and configuration to use (default: $HOME/.defradb)
      --schema string        Collection schema Root
      --tx uint              Transaction ID
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
      --version string       Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

