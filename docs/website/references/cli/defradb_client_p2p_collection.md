## defradb client p2p collection

Configure the P2P collection system

### Synopsis

Add, delete, or get the list of P2P collections.
The selected collections synchronize their events on the pubsub network.

### Options

```
  -h, --help   help for collection
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

* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client p2p collection add](defradb_client_p2p_collection_add.md)	 - Add P2P collections
* [defradb client p2p collection getall](defradb_client_p2p_collection_getall.md)	 - Get all P2P collections
* [defradb client p2p collection remove](defradb_client_p2p_collection_remove.md)	 - Remove P2P collections

