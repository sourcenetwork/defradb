## defradb client collection docIDs

List all document IDs (docIDs).

### Synopsis

List all document IDs (docIDs).
		
Example:
  defradb client collection docIDs --name User
		

```
defradb client collection docIDs [flags]
```

### Options

```
  -h, --help   help for docIDs
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
      --get-inactive                  Get inactive collections as well as active
      --logformat string              Log format to use. Options are csv, json (default "csv")
      --loglevel string               Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor                    Disable colored log output
      --logoutput string              Log output path (default "stderr")
      --logtrace                      Include stacktrace in error and fatal logs
      --max-txn-retries int           Specify the maximum number of retries per transaction (default 5)
      --name string                   Collection name
      --no-p2p                        Disable the peer-to-peer network synchronization system
      --p2paddr strings               Listen addresses for the p2p network (formatted as a libp2p MultiAddr) (default [/ip4/127.0.0.1/tcp/9171])
      --peers stringArray             List of peers to connect to
      --privkeypath string            Path to the private key for tls
      --pubkeypath string             Path to the public key for tls
      --rootdir string                Directory for persistent data (default: $HOME/.defradb)
      --schema string                 Collection schema Root
      --store string                  Specify the datastore to use (supported: badger, memory) (default "badger")
      --tx uint                       Transaction ID
      --url string                    URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --valuelogfilesize int          Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
      --version string                Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

