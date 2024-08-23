## defradb keyring

Manage DefraDB private keys

### Synopsis

Manage DefraDB private keys.
Generate, import, and export private keys.

The following keys are loaded from the keyring on start:
	peer-key: Ed25519 private key (required)
	encryption-key: AES-128, AES-192, or AES-256 key (optional)

To randomly generate the required keys, run the following command:
	defradb keyring generate

To import externally generated keys, run the following command:
	defradb keyring import <name> <private-key-hex>

To learn more about the available options:
	defradb keyring --help


### Options

```
  -h, --help   help for keyring
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

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb keyring export](defradb_keyring_export.md)	 - Export a private key
* [defradb keyring generate](defradb_keyring_generate.md)	 - Generate private keys
* [defradb keyring import](defradb_keyring_import.md)	 - Import a private key

