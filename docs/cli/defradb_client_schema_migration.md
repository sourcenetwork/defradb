## defradb client schema migration

Interact with the schema migration system of a running DefraDB instance

### Synopsis

Make set or look for existing schema migrations on a DefraDB node.

### Options

```
  -h, --help   help for migration
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

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client schema migration get](defradb_client_schema_migration_get.md)	 - Gets the schema migrations within DefraDB
* [defradb client schema migration set](defradb_client_schema_migration_set.md)	 - Set a schema migration within DefraDB

