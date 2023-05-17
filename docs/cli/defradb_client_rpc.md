## defradb client rpc

Interact with a DefraDB gRPC server

### Synopsis

Interact with a DefraDB gRPC server.

### Options

```
      --addr string   gRPC endpoint address (default "0.0.0.0:9161")
  -h, --help          help for rpc
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
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a running DefraDB node as a client
* [defradb client rpc p2pcollection](defradb_client_rpc_p2pcollection.md)	 - Interact with the P2P collection system
* [defradb client rpc replicator](defradb_client_rpc_replicator.md)	 - Interact with the replicator system

