## defradb client p2p replicator set

Add replicator(s) and start synchronization

### Synopsis

Add replicator(s) and start synchronization.
A replicator synchronizes one or all collection(s) from this node to another.

Example:
  defradb client p2p replicator set -c Users '{"ID": "12D3", "Addrs": ["/ip4/0.0.0.0/tcp/9171"]}'


```
defradb client p2p replicator set [-c, --collection] <peer> [flags]
```

### Options

```
  -c, --collection strings   Collection(s) to replicate
  -h, --help                 help for set
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
  -i, --identity string               ACP Identity
      --log-format string             Log format to use. Options are text or json (default "text")
      --log-level string              Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string             Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string          Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                    Include source location in logs
      --log-stacktrace                Include stacktrace in error and fatal logs
      --max-txn-retries int           Specify the maximum number of retries per transaction (default 5)
      --no-p2p                        Disable the peer-to-peer network synchronization system
      --p2paddr strings               Listen addresses for the p2p network (formatted as a libp2p MultiAddr) (default [/ip4/127.0.0.1/tcp/9171])
      --peers stringArray             List of peers to connect to
      --privkeypath string            Path to the private key for tls
      --pubkeypath string             Path to the public key for tls
      --rootdir string                Directory for persistent data (default: $HOME/.defradb)
      --store string                  Specify the datastore to use (supported: badger, memory) (default "badger")
      --tx uint                       Transaction ID
      --url string                    URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --valuelogfilesize int          Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
```

### SEE ALSO

* [defradb client p2p replicator](defradb_client_p2p_replicator.md)	 - Configure the replicator system

