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

* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client p2p replicator delete](defradb_client_p2p_replicator_delete.md)	 - Delete replicator(s) and stop synchronization
* [defradb client p2p replicator getall](defradb_client_p2p_replicator_getall.md)	 - Get all replicators
* [defradb client p2p replicator set](defradb_client_p2p_replicator_set.md)	 - Add replicator(s) and start synchronization

