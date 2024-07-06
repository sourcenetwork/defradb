## defradb client schema patch

Patch an existing schema type

### Synopsis

Patch an existing schema.

Uses JSON Patch to modify schema types.

Example: patch from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]' '{"lenses": [...'

Example: patch from file:
  defradb client schema patch -p patch.json

Example: patch from stdin:
  cat patch.json | defradb client schema patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.

```
defradb client schema patch [schema] [migration] [flags]
```

### Options

```
  -h, --help                help for patch
  -t, --lens-file string    File to load a lens config from
  -p, --patch-file string   File to load a patch from
      --set-active          Set the active schema version for all collections using the root schem
```

### Options inherited from parent commands

```
  -i, --identity string            Hex formatted private key used to authenticate with ACP
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --no-log-color               Disable colored log output
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node

