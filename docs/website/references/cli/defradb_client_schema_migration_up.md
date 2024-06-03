## defradb client schema migration up

Applies the migration to the specified collection version.

### Synopsis

Applies the migration to the specified collection version.
Documents is a list of documents to apply the migration to.		

Example: migrate from string
  defradb client schema migration up --collection 2 '[{"name": "Bob"}]'

Example: migrate from file
  defradb client schema migration up --collection 2 -f documents.json

Example: migrate from stdin
  cat documents.json | defradb client schema migration up --collection 2 -
		

```
defradb client schema migration up --collection <collectionID> <documents> [flags]
```

### Options

```
      --collection uint32   Collection id
  -f, --file string         File containing document(s)
  -h, --help                help for up
```

### Options inherited from parent commands

```
  -i, --identity string            Hex formatted private key used to authenticate with ACP
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

* [defradb client schema migration](defradb_client_schema_migration.md)	 - Interact with the schema migration system of a running DefraDB instance

