## defradb start

Start a DefraDB node

### Synopsis

Start a DefraDB node.

```
defradb start [flags]
```

### Options

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
      --development                   Enables a set of features that make development easier but should not be enabled in production:
                                       - allows purging of all persisted data 
                                       - generates temporary node identity if keyring is disabled
  -h, --help                          help for start
      --max-txn-retries int           Specify the maximum number of retries per transaction (default 5)
      --no-encryption                 Skip generating an encryption key. Encryption at rest will be disabled. WARNING: This cannot be undone.
      --no-p2p                        Disable the peer-to-peer network synchronization system
      --p2paddr strings               Listen addresses for the p2p network (formatted as a libp2p MultiAddr) (default [/ip4/127.0.0.1/tcp/9171])
      --peers stringArray             List of peers to connect to
      --privkeypath string            Path to the private key for tls
      --pubkeypath string             Path to the public key for tls
      --store string                  Specify the datastore to use (supported: badger, memory) (default "badger")
      --valuelogfilesize int          Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
```

### Options inherited from parent commands

```
      --keyring-backend string      Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string    Service name to use when using the system backend (default "defradb")
      --keyring-path string         Path to store encrypted keys when using the file backend (default "keys")
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

