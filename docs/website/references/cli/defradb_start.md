## defradb start

Start a DefraDB node

### Synopsis

Start a DefraDB node.

```
defradb start [flags]
```

### Options

```
      --allowed-origins stringArray       List of origins to allow for CORS requests. Their hosts are also accepted as auth token audiences
      --audience string                   Audience to set on minted auth tokens. Defaults to the host of --url
      --default-key-type string           Default key type to generate new node identity if one doesn't exist in the keyring. Valid values are 'secp256k1' and 'ed25519'. If not specified, the default key type will be 'secp256k1'. (default "secp256k1")
      --development                       Enables a set of features that make development easier but should not be enabled in production:
                                           - allows purging of all persisted data
                                           - generates temporary node identity if one doesn't exist in the keyring
      --document-acp-type string          Specify the document acp engine to use (supported: local (default), source-hub) (default "local")
  -h, --help                              help for start
  -i, --identity string                   Hex formatted private key used to authenticate with ACP
      --keyring-backend string            Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string          Service name to use when using the system backend (default "defradb")
      --keyring-path string               Path (relative to DefraDB root directory) to store encrypted keys when using the file backend (default "keys")
      --max-txn-retries int               Specify the maximum number of retries per transaction (default 5)
      --no-encryption                     Skip generating an encryption key. Encryption at rest will be disabled. WARNING: This cannot be undone.
      --no-keyring                        Disable the keyring and generate ephemeral keys
      --no-p2p                            Disable the peer-to-peer network synchronization system
      --no-searchable-encryption          Skip generating a searchable encryption key. Searchable encryption will be disabled.
      --no-signing                        Disable signing of commits.
      --no-telemetry                      Disables telemetry reporting. Telemetry is only enabled in builds that use the telemetry flag.
      --node-acp-enable                   Enable the node access control system.
      --p2paddr strings                   Listen addresses for the p2p network (formatted as a libp2p MultiAddr) (default [/ip4/127.0.0.1/tcp/9171])
      --peers stringArray                 List of peers to connect to
      --privkeypath string                Path to the private key for tls
      --pubkeypath string                 Path to the public key for tls
      --pubsub                            Enable the pubsub system (default true)
      --relay                             Enable the p2p relay
      --replicator-retry-intervals ints   Retry intervals for the replicator. Format is a comma-separated list of whole number seconds. Example: 10,20,40,80,160,320 (default [30,60,120,240,480,960,1920])
      --secret-file string                Path to the file containing secrets (default ".env")
      --source-hub-address string         The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --store string                      Specify the datastore to use (supported: badger, memory) (default "badger")
      --url string                        URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --valuelogfilesize int              Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
```

### Options inherited from parent commands

```
      --log-format string      Log format to use. Options are text or json (default "text")
      --log-level string       Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string      Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string   Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source             Include source location in logs
      --log-stacktrace         Include stacktrace in error and fatal logs
      --no-log-color           Disable colored log output
      --rootdir string         Directory for persistent data (default: $HOME/.defradb)
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

