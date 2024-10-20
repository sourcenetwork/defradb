## defradb client p2p replicator

Configure the replicator system

### Synopsis

Configure the replicator system. Add, delete, or get the list of persisted replicators.
A replicator replicates one or all collection(s) from one node to another.

### Options

```
  -h, --help   help for replicator
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

* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client p2p replicator delete](defradb_client_p2p_replicator_delete.md)	 - Delete replicator(s) and stop synchronization
* [defradb client p2p replicator getall](defradb_client_p2p_replicator_getall.md)	 - Get all replicators
* [defradb client p2p replicator set](defradb_client_p2p_replicator_set.md)	 - Add replicator(s) and start synchronization

