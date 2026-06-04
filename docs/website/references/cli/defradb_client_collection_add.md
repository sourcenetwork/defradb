## defradb client collection add

Add new collection

### Synopsis

Add new collection.

Collection type with a '@policy(id:".." resource: "..")' linked will only be accepted if:
  - ACP is available (i.e. ACP is not disabled).
  - The specified resource adheres to the document resource interface (DRI).
  - Learn more about the DefraDB [ACP System](https://docs.source.network/defradb/references/acp)

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client collection add [sdl] [flags]
```

### Examples

```
add from an argument string:  
  defradb client collection add 'type Foo { ... }'

add from file:  
  defradb client collection add -f schema.graphql

add from multiple files:  
  defradb client collection add -f schema1.graphql -f schema2.graphql

add from multiple files (comma-separated):  
  defradb client collection add -f schema1.graphql,schema2.graphql

add from stdin:  
  cat schema.graphql | defradb client collection add -
```

### Options

```
  -f, --file strings   File to load a collection definition from
  -h, --help           help for add
```

### Options inherited from parent commands

```
      --collection-id string        Collection ID
      --get-inactive                Get inactive collections as well as active
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
      --name string                 Collection name
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version-id string           Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

