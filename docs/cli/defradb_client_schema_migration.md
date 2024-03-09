## defradb client schema migration

Interact with the schema migration system of a running DefraDB instance

### Synopsis

Make set or look for existing schema migrations on a DefraDB node.

### Options

```
  -h, --help   help for migration
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

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client schema migration down](defradb_client_schema_migration_down.md)	 - Reverses the migration to the specified collection version.
* [defradb client schema migration reload](defradb_client_schema_migration_reload.md)	 - Reload the schema migrations within DefraDB
* [defradb client schema migration set](defradb_client_schema_migration_set.md)	 - Set a schema migration within DefraDB
* [defradb client schema migration set-registry](defradb_client_schema_migration_set-registry.md)	 - Set a schema migration within the DefraDB LensRegistry
* [defradb client schema migration up](defradb_client_schema_migration_up.md)	 - Applies the migration to the specified collection version.

