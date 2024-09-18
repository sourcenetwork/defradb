## defradb identity

Interact with identity features of DefraDB instance

### Synopsis

Interact with identity features of DefraDB instance

### Options

```
  -h, --help   help for identity
```

### Options inherited from parent commands

```
      --keyring-backend string       Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string     Service name to use when using the system backend (default "defradb")
      --keyring-path string          Path to store encrypted keys when using the file backend (default "keys")
      --keyring-secret-file string   Path to the file containing the keyring secret (default ".env")
      --log-format string            Log format to use. Options are text or json (default "text")
      --log-level string             Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string            Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string         Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                   Include source location in logs
      --log-stacktrace               Include stacktrace in error and fatal logs
      --no-keyring                   Disable the keyring and generate ephemeral keys
      --no-log-color                 Disable colored log output
      --rootdir string               Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string    The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --url string                   URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb identity new](defradb_identity_new.md)	 - Generate a new identity

