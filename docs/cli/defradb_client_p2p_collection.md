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
* [defradb client p2p collection add](defradb_client_p2p_collection_add.md)	 - Add P2P collections
* [defradb client p2p collection getall](defradb_client_p2p_collection_getall.md)	 - Get all P2P collections
* [defradb client p2p collection remove](defradb_client_p2p_collection_remove.md)	 - Remove P2P collections

