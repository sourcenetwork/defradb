## defradb client collection patch

Patch existing collection descriptions

### Synopsis

Patch existing collection descriptions.

Uses JSON Patch to modify collection descriptions.

Example: patch from an argument string:
  defradb client collection patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: patch from file:
  defradb client collection patch -p patch.json

Example: patch from stdin:
  cat patch.json | defradb client collection patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.

```
defradb client collection patch [patch] [flags]
```

### Options

```
  -h, --help                help for patch
  -p, --patch-file string   File to load a patch from
```

### Options inherited from parent commands

```
      --get-inactive               Get inactive collections as well as active
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
      --name string                Collection name
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --schema string              Collection schema Root
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version string             Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

