## defradb client collection

View detailed collection info.

### Synopsis

View detailed collection info.
		
Example: view all collections
  defradb client collection

Example: view collection by name
  defradb client collection --name User

Example: view collection by schema id
  defradb client collection --schema bae123

Example: view collection by version id
  defradb client collection --version bae123
		

```
defradb client collection [--name <name> --schema <schemaID> --version <versionID>] [flags]
```

### Options

```
  -h, --help             help for collection
      --name string      Get collection by name
      --schema string    Get collection by schema ID
      --version string   Get collection by version ID
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

