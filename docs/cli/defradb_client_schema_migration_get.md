## defradb client schema migration get

Gets the schema migrations within DefraDB

### Synopsis

Gets the schema migrations within the local DefraDB node.

Example:
  defradb client schema migration get'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client schema migration get [flags]
```

### Options

```
  -h, --help   help for get
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
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client schema migration](defradb_client_schema_migration.md)	 - Interact with the schema migration system of a running DefraDB instance

