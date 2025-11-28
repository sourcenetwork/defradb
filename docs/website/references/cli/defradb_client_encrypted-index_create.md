## defradb client encrypted-index create

Creates an encrypted index on a collection's field

### Synopsis

Creates an encrypted index on a collection's field.

The --type flag is optional. If not provided, the default value will be "equality".

Currently only "equality" type is supported.

```
defradb client encrypted-index create -c --collection <collection> --field <field> [--type <type>] [flags]
```

### Examples

```
create an index for 'Users' collection on 'name' field:  
  defradb client encrypted-index create --collection Users --field name
```

### Options

```
  -c, --collection string   Collection name
      --field string        Field to index
  -h, --help                help for create
      --type string         Type of index to create
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

* [defradb client encrypted-index](defradb_client_encrypted-index.md)	 - Manage collections' encrypted indexes on a running DefraDB node

