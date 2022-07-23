## defradb start

Start a DefraDB node

### Synopsis

Start a new instance of DefraDB node.

```
defradb start [flags]
```

### Options

```
  -h, --help             help for start
      --no-p2p           Disable the peer-to-peer network synchronization system
      --p2paddr string   Listener address for the p2p network (formatted as a libp2p MultiAddr) (default "/ip4/0.0.0.0/tcp/9171")
      --peers string     List of peers to connect to
      --store string     Specify the datastore to use (supported: badger, memory) (default "badger")
      --tcpaddr string   Listener address for the tcp gRPC server (formatted as a libp2p MultiAddr) (default "/ip4/0.0.0.0/tcp/9161")
```

### Options inherited from parent commands

```
      --logcolor           Enable colored output
      --logformat string   Log format to use. Options are text, json (default "csv")
      --loglevel string    Log level to use. Options are debug, info, error, fatal (default "info")
      --logoutput string   Log output path (default "stderr")
      --logtrace           Include stacktrace in error and fatal logs
      --rootdir string     Directory for data and configuration to use (default "$HOME/.defradb")
      --url string         URL of the target database's HTTP endpoint (default "localhost:9181")
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

