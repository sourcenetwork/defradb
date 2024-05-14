## defradb client index

Manage collections' indexes of a running DefraDB instance

### Synopsis

Manage (create, drop, or list) collection indexes on a DefraDB node.

### Options

```
  -h, --help   help for index
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
* [defradb client index create](defradb_client_index_create.md)	 - Creates a secondary index on a collection's field(s)
* [defradb client index drop](defradb_client_index_drop.md)	 - Drop a collection's secondary index
* [defradb client index list](defradb_client_index_list.md)	 - Shows the list indexes in the database or for a specific collection

