## defradb client schema

Interact with the schema system of a DefraDB node

### Synopsis

Make changes, updates, or look for existing schema types.

### Options

```
  -h, --help   help for schema
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
* [defradb client schema add](defradb_client_schema_add.md)	 - Add new schema
* [defradb client schema describe](defradb_client_schema_describe.md)	 - View schema descriptions.
* [defradb client schema migration](defradb_client_schema_migration.md)	 - Interact with the schema migration system of a running DefraDB instance
* [defradb client schema patch](defradb_client_schema_patch.md)	 - Patch an existing schema type
* [defradb client schema set-default](defradb_client_schema_set-default.md)	 - Set the default schema version

