## defradb client collection patch

Patch existing collection versions

### Synopsis

Patch existing collection versions.

Uses JSON Patch to modify collection versions.

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.

```
defradb client collection patch [patch] [migration] [flags]
```

### Examples

```
patch from an argument string:  
  defradb client collection patch '[{ "op": "add", "path": "...", "value": {...} }]' '{"lenses": [...'

patch from file:  
  defradb client collection patch -p patch.json

patch from stdin:  
  cat patch.json | defradb client collection patch -
```

### Options

```
  -h, --help                help for patch
  -t, --lens-file string    File to load a lens config from
  -p, --patch-file string   File to load a patch from
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

