## defradb client rpc replicator

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
      --addr string          RPC endpoint address (default "0.0.0.0:9161")
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

* [defradb client rpc](defradb_client_rpc.md)	 - Interact with a DefraDB node via RPC
* [defradb client rpc replicator delete](defradb_client_rpc_replicator_delete.md)	 - Delete a replicator. It will stop synchronizing
* [defradb client rpc replicator getall](defradb_client_rpc_replicator_getall.md)	 - Get all replicators
* [defradb client rpc replicator set](defradb_client_rpc_replicator_set.md)	 - Set a P2P replicator

