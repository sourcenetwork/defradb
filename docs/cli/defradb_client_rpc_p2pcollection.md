## defradb client rpc p2pcollection

Configure the P2P collection system

### Synopsis

Add, delete, or get the list of P2P collections.
The selected collections synchronize their events on the pubsub network.

### Options

```
  -h, --help   help for p2pcollection
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
* [defradb client rpc p2pcollection add](defradb_client_rpc_p2pcollection_add.md)	 - Add P2P collections
* [defradb client rpc p2pcollection getall](defradb_client_rpc_p2pcollection_getall.md)	 - Get all P2P collections
* [defradb client rpc p2pcollection remove](defradb_client_rpc_p2pcollection_remove.md)	 - Remove P2P collections

