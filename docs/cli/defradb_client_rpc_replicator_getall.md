## defradb client rpc replicator getall

Get all replicators

### Synopsis

Get all the replicators active in the P2P data sync system.
These are the replicators that are currently replicating data from one node to another.

```
defradb client rpc replicator getall [flags]
```

### Options

```
  -h, --help   help for getall
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

* [defradb client rpc replicator](defradb_client_rpc_replicator.md)	 - Configure the replicator system

