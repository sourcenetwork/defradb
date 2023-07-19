## defradb client rpc replicator

Interact with the replicator system

### Synopsis

Add, delete, or get the list of persisted replicators

### Options

```
  -h, --help   help for replicator
```

### Options inherited from parent commands

```
      --addr string          gRPC endpoint address (default "0.0.0.0:9161")
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --rootdir string       Directory for data and configuration to use (default: $HOME/.defradb)
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client rpc](defradb_client_rpc.md)	 - Interact with a DefraDB gRPC server
* [defradb client rpc replicator delete](defradb_client_rpc_replicator_delete.md)	 - Delete a replicator
* [defradb client rpc replicator getall](defradb_client_rpc_replicator_getall.md)	 - Get all replicators
* [defradb client rpc replicator set](defradb_client_rpc_replicator_set.md)	 - Set a P2P replicator

