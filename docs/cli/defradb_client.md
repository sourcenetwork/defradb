## defradb client

Interact with a DefraDB node

### Synopsis

Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.

### Options

```
  -h, --help              help for client
  -i, --identity string   ACP Identity
      --tx uint           Transaction ID
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
      --keyring-path string           Path to store encrypted keys (default "keys")
      --log-format string             Log format to use. Options are text or json (default "text")
      --log-level string              Log level to use. Options are debug, info, error, fatal (default "info")
      --log-no-color                  Disable colored log output
      --log-output string             Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string          Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                    Include source location in logs
      --log-stacktrace                Include stacktrace in error and fatal logs
      --max-txn-retries int           Specify the maximum number of retries per transaction (default 5)
      --no-keyring                    Disable the keyring and generate ephemeral keys
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

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb client acp](defradb_client_acp.md)	 - Interact with the access control system of a DefraDB node
* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility
* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.
* [defradb client dump](defradb_client_dump.md)	 - Dump the contents of DefraDB node-side
* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance
* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client query](defradb_client_query.md)	 - Send a DefraDB GraphQL query request
* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client tx](defradb_client_tx.md)	 - Create, commit, and discard DefraDB transactions
* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

