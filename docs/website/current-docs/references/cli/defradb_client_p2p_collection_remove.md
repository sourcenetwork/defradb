## defradb client p2p collection remove

Remove P2P collections

### Synopsis

Remove P2P collections from the followed pubsub topics.
The removed collections will no longer be synchronized between nodes.

Example: remove single collection
  defradb client p2p collection remove bae123

Example: remove multiple collections
  defradb client p2p collection remove bae123,bae456
		

```
defradb client p2p collection remove [collectionIDs] [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
  -i, --identity string            Hex formatted private key used to authenticate with ACP
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
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client p2p collection](defradb_client_p2p_collection.md)	 - Configure the P2P collection system

