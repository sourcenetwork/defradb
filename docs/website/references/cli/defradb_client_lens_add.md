## defradb client lens add

Add a lens to the lens store

### Synopsis

Add a lens configuration to the lens store and return its CID.

The lens store is content-addressed, so identical lens configurations
will return the same CID without duplicating storage.

```
defradb client lens add [cfg] [flags]
```

### Examples

```
add from an argument string:  
  defradb client lens add '{"lenses": [...'

add from file:  
  defradb client lens add -f lens_config.json

add from stdin:  
  cat lens_config.json | defradb client lens add -
```

### Options

```
  -f, --file string   Lens configuration file
  -h, --help          help for add
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

* [defradb client lens](defradb_client_lens.md)	 - Interact with the collection migration system of a running DefraDB instance

