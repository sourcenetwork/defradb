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
  -f, --file string   File to load a policy from
  -h, --help          help for add
```

### Options inherited from parent commands

```
  -i, --identity string            ACP Identity
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-no-color               Disable colored log output
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client acp policy](defradb_client_acp_policy.md)	 - Interact with the acp policy features of DefraDB instance

