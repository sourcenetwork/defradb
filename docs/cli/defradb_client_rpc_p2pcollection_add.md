## defradb client rpc p2pcollection add

Add p2p collections

### Synopsis

Add p2p collections to the synchronized pubsub topics.
The collections are synchronized between nodes of a pubsub network.

```
defradb client rpc p2pcollection add [collectionID] [flags]
```

### Options

```
  -h, --help   help for add
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

* [defradb client rpc p2pcollection](defradb_client_rpc_p2pcollection.md)	 - Configure the p2p collection system

