## defradb client collection get

View document fields.

### Synopsis

View document fields.

Example:
  defradb client collection get --name User bae-123
		

```
defradb client collection get <docKey> [--show-deleted] [flags]
```

### Options

```
  -h, --help           help for get
      --show-deleted   Show deleted documents
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

