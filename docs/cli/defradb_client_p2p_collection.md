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
      --allowed-origins stringArray   List of origins to allow for CORS requests
      --logformat string              Log format to use. Options are csv, json (default "csv")
      --loglevel string               Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor                    Disable colored log output
      --logoutput string              Log output path (default "stderr")
      --logtrace                      Include stacktrace in error and fatal logs
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

* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client p2p collection add](defradb_client_p2p_collection_add.md)	 - Add P2P collections
* [defradb client p2p collection getall](defradb_client_p2p_collection_getall.md)	 - Get all P2P collections
* [defradb client p2p collection remove](defradb_client_p2p_collection_remove.md)	 - Remove P2P collections

