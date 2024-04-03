## defradb client acp policy add

Add new policy

### Synopsis

Add new policy

Notes:
  - Can not add a policy without specifying an identity.
  - ACP must be available (i.e. ACP can not be disabled).
  - A non-DPI policy will be accepted (will be registered with acp system).
  - But only a valid DPI policyID & resource can be specified on a schema.
  - DPI validation happens when attempting to add a schema with '@policy'.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: add from an argument string:
  defradb client acp policy add -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j '
description: A Valid DefraDB Policy Interface

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
'

Example: add from file:
  defradb client acp policy add -i cosmos17r39df0hdcrgnmmw4mvu7qgk5nu888c7uvv37y -f policy.yml

Example: add from file, verbose flags:
  defradb client acp policy add --identity cosmos1kpw734v54g0t0d8tcye8ee5jc3gld0tcr2q473 --file policy.yml

Example: add from stdin:
  cat policy.yml | defradb client acp policy add -



```
defradb client acp policy add [-i --identity] [policy] [flags]
```

### Options

```
  -f, --file string       File to load a policy from
  -h, --help              help for add
  -i, --identity string   [Required] Identity of the creator
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

* [defradb client acp policy](defradb_client_acp_policy.md)	 - Interact with the acp policy features of DefraDB instance

