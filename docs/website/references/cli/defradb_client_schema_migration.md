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
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --keyring-backend string      Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string    Service name to use when using the system backend (default "defradb")
      --keyring-path string         Path to store encrypted keys when using the file backend (default "keys")
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client schema migration down](defradb_client_schema_migration_down.md)	 - Reverses the migration to the specified collection version.
* [defradb client schema migration reload](defradb_client_schema_migration_reload.md)	 - Reload the schema migrations within DefraDB
* [defradb client schema migration set](defradb_client_schema_migration_set.md)	 - Set a schema migration within DefraDB
* [defradb client schema migration set-registry](defradb_client_schema_migration_set-registry.md)	 - Set a schema migration within the DefraDB LensRegistry
* [defradb client schema migration up](defradb_client_schema_migration_up.md)	 - Applies the migration to the specified collection version.

