## defradb

DefraDB Edge Database

### Synopsis

DefraDB is the edge database to power the user-centric future.

Start a DefraDB node, interact with a local or remote node, and much more.


### Options

```
  -h, --help                       help for defradb
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
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb identity](defradb_identity.md)	 - Interact with identity features of DefraDB instance
* [defradb keyring](defradb_keyring.md)	 - Manage DefraDB private keys
* [defradb server-dump](defradb_server-dump.md)	 - Dumps the state of the entire database
* [defradb start](defradb_start.md)	 - Start a DefraDB node
* [defradb version](defradb_version.md)	 - Display the version information of DefraDB and its components

