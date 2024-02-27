## defradb client collection delete

Delete documents by docID or filter.

### Synopsis

Delete documents by docID or filter and lists the number of documents deleted.
		
Example: delete by docID(s):
  defradb client collection delete  --name User --docID bae-123,bae-456

Example: delete by docID(s) with identity:
  defradb client collection delete -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j --name User --docID bae-123,bae-456

Example: delete by filter:
  defradb client collection delete --name User --filter '{ "_gte": { "points": 100 } }'
		

```
defradb client collection delete [-i --identity] [--filter <filter> --docID <docID>] [flags]
```

### Options

```
      --docID strings     Document ID
      --filter string     Document filter
  -h, --help              help for delete
  -i, --identity string   Identity of the actor
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
      --get-inactive                  Get inactive collections as well as active
      --log-format string             Log format to use. Options are text or json (default "text")
      --log-level string              Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string             Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string          Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                    Include source location in logs
      --log-stacktrace                Include stacktrace in error and fatal logs
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

