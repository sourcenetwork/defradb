## defradb client lens set

Set a schema migration within DefraDB

### Synopsis

Set a migration from a source schema version to a destination schema version for
all collections that are on the given source schema version within the local DefraDB node.

Example: set from an argument string:
  defradb client lens set bae123 bae456 '{"lenses": [...'

Example: set from file:
  defradb client lens set bae123 bae456 -f schema_migration.lens

Example: add from stdin:
  cat schema_migration.lens | defradb client lens set bae123 bae456 -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client lens set [src] [dst] [cfg] [flags]
```

### Options

```
  -f, --file string   Lens configuration file
  -h, --help          help for set
```

### Options inherited from parent commands

```
  -i, --identity string                     Hex formatted private key used to authenticate with ACP
      --keyring-backend string              Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string            Service name to use when using the system backend (default "defradb")
      --keyring-path string                 Path to store encrypted keys when using the file backend (default "keys")
      --log-format string                   Log format to use. Options are text or json (default "text")
      --log-level string                    Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string                   Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string                Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                          Include source location in logs
      --log-stacktrace                      Include stacktrace in error and fatal logs
      --no-keyring                          Disable the keyring and generate ephemeral keys
      --no-log-color                        Disable colored log output
      --replicator-retry-intervals string   Retry intervals for the replicator. Format is a comma-separated list of durations. Example: "10,20,40,80,160,320" (default "30,60,120,240,480,960,1920")
      --rootdir string                      Directory for persistent data (default: $HOME/.defradb)
      --secret-file string                  Path to the file containing secrets (default ".env")
      --source-hub-address string           The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                             Transaction ID
      --url string                          URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client lens](defradb_client_lens.md)	 - Interact with the schema migration system of a running DefraDB instance

