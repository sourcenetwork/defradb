## defradb client schema migration down

Reverses the migration from the specified schema version.

### Synopsis

Reverses the migration from the specified schema version.
Documents is a list of documents to reverse the migration from.

Example: migrate from string
  defradb client schema migration down --version bae123 '[{"name": "Bob"}]'

Example: migrate from file
  defradb client schema migration down --version bae123 -f documents.json

Example: migrate from stdin
  cat documents.json | defradb client schema migration down --version bae123 -
		

```
defradb client schema migration down --version <version> <documents> [flags]
```

### Options

```
  -f, --file string      File containing document(s)
  -h, --help             help for down
      --version string   Schema version id
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

