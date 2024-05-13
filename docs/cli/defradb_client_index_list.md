## defradb client index list

Shows the list indexes in the database or for a specific collection

### Synopsis

Shows the list indexes in the database or for a specific collection.
		
If the --collection flag is provided, only the indexes for that collection will be shown.
Otherwise, all indexes in the database will be shown.

Example: show all index for 'Users' collection:
  defradb client index list --collection Users

```
defradb client index list [-c --collection <collection>] [flags]
```

### Options

```
  -c, --collection string   Collection name
  -h, --help                help for list
```

### Options inherited from parent commands

```
  -i, --identity string            ACP Identity
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-no-color               Disable colored log output
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance

