## defradb client collection

Interact with a collection.

### Synopsis

Create, read, update, and delete documents within a collection.

### Options

```
      --get-inactive     Get inactive collections as well as active
  -h, --help             help for collection
      --name string      Collection name
      --schema string    Collection schema Root
      --tx uint          Transaction ID
      --version string   Collection version ID
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
      --url string                    URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --valuelogfilesize int          Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client collection create](defradb_client_collection_create.md)	 - Create a new document.
* [defradb client collection delete](defradb_client_collection_delete.md)	 - Delete documents by docID or filter.
* [defradb client collection describe](defradb_client_collection_describe.md)	 - View collection description.
* [defradb client collection docIDs](defradb_client_collection_docIDs.md)	 - List all document IDs (docIDs).
* [defradb client collection get](defradb_client_collection_get.md)	 - View document fields.
* [defradb client collection update](defradb_client_collection_update.md)	 - Update documents by docID or filter.

