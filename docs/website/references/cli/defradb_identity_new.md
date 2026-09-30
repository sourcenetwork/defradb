## defradb identity new

Generate a new identity

### Synopsis

Generate a new identity

The generated identity contains:
- A private key (secp256k1 or ed25519, based on --type flag)
- A corresponding public key
- A "did:key" generated from the public key.

```
defradb identity new [flags]
```

### Examples

```
generate a new identity with secp256k1 key (default):  
  defradb identity new

generate a new identity with ed25519 key:  
  defradb identity new --type ed25519
```

### Options

```
  -h, --help          help for new
      --type string   Key type (secp256k1 or ed25519) (default "secp256k1")
```

### Options inherited from parent commands

```
      --log-format string      Log format to use. Options are text or json (default "text")
      --log-level string       Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string      Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string   Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source             Include source location in logs
      --log-stacktrace         Include stacktrace in error and fatal logs
      --no-log-color           Disable colored log output
      --rootdir string         Directory for persistent data (default: $HOME/.defradb)
```

### SEE ALSO

* [defradb identity](defradb_identity.md)	 - Interact with identity features of DefraDB instance

