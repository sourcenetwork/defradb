## defradb client schema add

Add new schema

### Synopsis

Add new schema.

Schema Object with a '@policy(id:".." resource: "..")' linked will only be accepted if:
  - ACP is available (i.e. ACP is not disabled).
  - The specified resource adheres to the Document Access Control DPI Rules.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: add from an argument string:
  defradb client schema add 'type Foo { ... }'

Example: add from file:
  defradb client schema add -f schema.graphql

Example: add from stdin:
  cat schema.graphql | defradb client schema add -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client schema add [schema] [flags]
```

### Options

```
  -f, --file string   File to load a schema from
  -h, --help          help for add
```

### Options inherited from parent commands

```
      --allowed-origins stringArray   List of origins to allow for CORS requests
  -i, --identity string               ACP Identity
      --keyring-backend string        Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string      Service name to use when using the system backend (default "defradb")
      --keyring-path string           Path to store encrypted keys when using the file backend (default "keys")
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
      --tx uint                       Transaction ID
      --url string                    URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --valuelogfilesize int          Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1073741824)
```

### SEE ALSO

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node

