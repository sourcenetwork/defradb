## defradb client document update

Update documents by key or filter.

### Synopsis

Update documents by key or filter.
		
Example:
  defradb client document update --collection User --key bae123 '{ "name": "Bob" }'

Example: update by filter
  defradb client document update --collection User \
  --filter '{ "_gte": { "points": 100 } }' --updater '{ "verified": true }'

Example: update by keys
  defradb client document update --collection User \
  --key bae123,bae456 --updater '{ "verified": true }'
		

```
defradb client document update --collection <collection> [--filter <filter> --key <key> --updater <updater>] <document> [flags]
```

### Options

```
  -c, --collection string   Collection name
      --filter string       Document filter
  -h, --help                help for update
      --key strings         Document key
      --updater string      Document updater
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

