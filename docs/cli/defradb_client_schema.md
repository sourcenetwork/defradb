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
  -i, --identity string            ACP Identity
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --no-log-color               Disable colored log output
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client schema add](defradb_client_schema_add.md)	 - Add new schema
* [defradb client schema describe](defradb_client_schema_describe.md)	 - View schema descriptions.
* [defradb client schema migration](defradb_client_schema_migration.md)	 - Interact with the schema migration system of a running DefraDB instance
* [defradb client schema patch](defradb_client_schema_patch.md)	 - Patch an existing schema type
* [defradb client schema set-active](defradb_client_schema_set-active.md)	 - Set the active collection version

