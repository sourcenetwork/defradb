## defradb keyring generate

Generate private keys

### Synopsis

Generate private keys.
Randomly generate and store private keys in the keyring.
By default peer and encryption keys will be generated.

WARNING: This will overwrite existing keys in the keyring.

Example:
  defradb keyring generate

Example: with no encryption key
  defradb keyring generate --no-encryption-key

Example: with no peer key
  defradb keyring generate --no-peer-key

Example: with system keyring
  defradb keyring generate --keyring-backend system

```
defradb keyring generate [flags]
```

### Options

```
  -h, --help                help for generate
      --no-encryption-key   Skip generating an encryption key. Encryption at rest will be disabled
      --no-peer-key         Skip generating a peer key.
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
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb keyring](defradb_keyring.md)	 - Manage DefraDB private keys

