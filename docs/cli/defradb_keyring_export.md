## defradb keyring export

Export a private key

### Synopsis

Export a private key.
Prints the hexadecimal representation of a private key.

Example:
  defradb keyring export encryption-key

```
defradb keyring export <name> [flags]
```

### Options

```
  -h, --help   help for export
```

### Options inherited from parent commands

```
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
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb keyring](defradb_keyring.md)	 - Manage DefraDB private keys

