## defradb client document delete

Delete documents by key or filter.

### Synopsis

Delete documents by key or filter and lists the number of documents deleted.
		
Example: delete by key(s)
  defradb client document delete --collection User --key bae123,bae456,...

Example: delete by filter
  defradb client document delete --collection User --filter '{ "_gte": { "points": 100 } }'
		

```
defradb client document delete --collection <collection> [--filter <filter> --key <key>] [flags]
```

### Options

```
  -c, --collection string   Collection name
      --filter string       Document filter
  -h, --help                help for delete
      --key strings         Document key
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

