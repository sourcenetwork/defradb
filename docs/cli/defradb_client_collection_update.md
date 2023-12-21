## defradb client collection update

Update documents by docID or filter.

### Synopsis

Update documents by docID or filter.
		
Example: update from string
  defradb client collection update --name User --docID bae-123 '{ "name": "Bob" }'

Example: update by filter
  defradb client collection update --name User \
  --filter '{ "_gte": { "points": 100 } }' --updater '{ "verified": true }'

Example: update by docIDs
  defradb client collection update --name User \
  --docID bae-123,bae-456 --updater '{ "verified": true }'
		

```
defradb client collection update [--filter <filter> --docID <docID> --updater <updater>] <document> [flags]
```

### Options

```
      --docID strings    Document ID
      --filter string    Document filter
  -h, --help             help for update
      --updater string   Document updater
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

