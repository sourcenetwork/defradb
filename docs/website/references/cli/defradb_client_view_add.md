## defradb client view add

Add new view

### Synopsis

Add new database view.

Use --lens-cid to specify a lens transform. Store a lens first using 'defradb client lens add

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client view add [query|query-file] [sdl|sdl-file] [flags]
```

### Examples

```
add a simple view from string flags:  
  defradb client view add --query 'Foo { name, ...}' --sdl 'type Foo { ... }'

add using an existing lens CID:  
  defradb client view add --query-file /path/to/query --sdl-file /path/to/sdl --lens-cid bafyreih...

add from file flags using an existing lens CID:  
  defradb client view add --query-file /path/to/query --sdl-file /path/to/sdl --lens-cid bafyreih...
```

### Options

```
  -h, --help                help for add
      --lens-cid string     CID of an existing lens transform (use 'lens add' first)
      --query string        Query
      --query-file string   Query file
      --sdl string          SDL
      --sdl-file string     SDL file
```

### Options inherited from parent commands

```
  -i, --identity string             Hex formatted private key used to authenticate with ACP
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
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

