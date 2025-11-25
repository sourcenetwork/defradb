## defradb client p2p collection sync-branchable

Synchronize a branchable collection's DAG from the network

### Synopsis

Synchronize a branchable collection's DAG from the network.

This command allows you to sync the collection-level history for branchable collections
(collections marked with @branchable directive). It doesn't automatically subscribe
to the collection for future updates.

```
defradb client p2p collection sync-branchable [collection-id] [flags]
```

### Examples

```
sync branchable collection:  
  defradb client p2p collection sync-branchable bafkreig27seqzxvr7isblvj77wvqnmkzoyv3u4nwytyethkbcpxlrx3iqq

sync branchable collection with timeout:  
  defradb client p2p collection sync-branchable bafkreig27seqzxvr7isblvj77wvqnmkzoyv3u4nwytyethkbcpxlrx3iqq --timeout 10s
```

### Options

```
  -h, --help               help for sync-branchable
      --timeout duration   Timeout for sync operations (default: 5s if not specified)
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

* [defradb client p2p collection](defradb_client_p2p_collection.md)	 - Configure the P2P collection system

