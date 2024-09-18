## defradb client index create

Creates a secondary index on a collection's field(s)

### Synopsis

Creates a secondary index on a collection's field(s).
		
The --name flag is optional. If not provided, a name will be generated automatically.
The --unique flag is optional. If provided, the index will be unique.

Example: create an index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name

Example: create a named index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name --name UsersByName

```
defradb client index create -c --collection <collection> --fields <fields> [-n --name <name>] [--unique] [flags]
```

### Options

```
  -c, --collection string   Collection name
      --fields strings      Fields to index
  -h, --help                help for create
  -n, --name string         Index name
  -u, --unique              Make the index unique
```

### Options inherited from parent commands

```
  -i, --identity string              Hex formatted private key used to authenticate with ACP
      --keyring-backend string       Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string     Service name to use when using the system backend (default "defradb")
      --keyring-path string          Path to store encrypted keys when using the file backend (default "keys")
      --keyring-secret-file string   Path to the file containing the keyring secret (default ".env")
      --log-format string            Log format to use. Options are text or json (default "text")
      --log-level string             Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string            Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string         Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                   Include source location in logs
      --log-stacktrace               Include stacktrace in error and fatal logs
      --no-keyring                   Disable the keyring and generate ephemeral keys
      --no-log-color                 Disable colored log output
      --rootdir string               Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string    The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                      Transaction ID
      --url string                   URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance

