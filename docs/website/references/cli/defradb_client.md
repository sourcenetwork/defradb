## defradb client

Interact with a DefraDB node

### Synopsis

Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.

### Options

```
  -h, --help              help for client
  -i, --identity string   Hex formatted private key used to authenticate with ACP
      --tx uint           Transaction ID
```

### Options inherited from parent commands

```
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
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb client acp](defradb_client_acp.md)	 - Interact with the access control system of a DefraDB node
* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility
* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.
* [defradb client dump](defradb_client_dump.md)	 - Dump the contents of DefraDB node-side
* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance
* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client purge](defradb_client_purge.md)	 - Delete all persisted data and restart
* [defradb client query](defradb_client_query.md)	 - Send a DefraDB GraphQL query request
* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client tx](defradb_client_tx.md)	 - Create, commit, and discard DefraDB transactions
* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

