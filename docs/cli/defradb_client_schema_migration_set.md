## defradb client schema migration set

Set a schema migration within DefraDB

### Synopsis

Set a migration between two schema versions within the local DefraDB node.

Example: set from an argument string:
  defradb client schema migration set bae123 bae456 '{"lenses": [...'

Example: set from file:
  defradb client schema migration set bae123 bae456 -f schema_migration.lens

Example: add from stdin:
  cat schema_migration.lens | defradb client schema migration set bae123 bae456 -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client schema migration set [src] [dst] [cfg] [flags]
```

### Options

```
  -f, --file string   Lens configuration file
  -h, --help          help for set
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

* [defradb client schema migration](defradb_client_schema_migration.md)	 - Interact with the schema migration system of a running DefraDB instance

