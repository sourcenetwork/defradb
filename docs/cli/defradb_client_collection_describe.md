## defradb client collection describe

View collection description.

### Synopsis

Introspect collection types.

Example: view all collections
  defradb client collection describe
		
Example: view collection by name
  defradb client collection describe --name User
		
Example: view collection by schema id
  defradb client collection describe --schema bae123
		
Example: view collection by version id
  defradb client collection describe --version bae123
		

```
defradb client collection describe [flags]
```

### Options

```
  -h, --help   help for describe
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

