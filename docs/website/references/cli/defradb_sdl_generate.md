## defradb sdl generate

Generate full GraphQL formatted schema.

### Synopsis

Generates the fully formatted GraphQL schema from a given user type definition(s).

Accepts multiple input files as well as "-" to use stdin.

```
defradb sdl generate --output schema.graphql <input schema files...> [flags]
```

### Examples

```
Generate SDL:  
  defradb sdl generate foo.graphql

Generate Multiple SDLs:  
  defradb sdl generate foo.graphql bar.graphql

Generate SDL and overwrite output:  
  defradb sdl generate foo.graphql bar.graphql --output schema.graphql -y

Generate SDL with Searchable Encryption type definitions:  
  defradb sdl generate foo.graphql -s
```

### Options

```
  -h, --help                            help for generate
  -s, --include-searchable-encryption   Include the schema type definitions to support Searchable Encryption
  -o, --output string                   The output file to write the generated schema. Accepts '-' to write to stdout (default "schema.gen.graphql")
  -y, --overwrite                       Overwrite any existing matching output file paths
```

### Options inherited from parent commands

```
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
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb sdl](defradb_sdl.md)	 - Utilities to interact with the DefraDB SDL

