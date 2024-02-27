## defradb client p2p collection remove

Remove P2P collections

### Synopsis

Remove P2P collections from the followed pubsub topics.
The removed collections will no longer be synchronized between nodes.

Example: remove single collection
  defradb client p2p collection remove bae123

Example: remove multiple collections
  defradb client p2p collection remove bae123,bae456
		

```
defradb client p2p collection remove [collectionIDs] [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
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

* [defradb client p2p collection](defradb_client_p2p_collection.md)	 - Configure the P2P collection system

