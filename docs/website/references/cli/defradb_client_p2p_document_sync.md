## defradb client p2p document sync

Synchronize specific documents from the network

### Synopsis

Synchronize specific documents from the network.

This command allows you to sync documents from a specific collection across the network.
It doesn't automatically subscribe to the collection or the documents.

```
defradb client p2p document sync <collection-name> <docID...> [flags]
```

### Examples

```
sync single document:  
  defradb client p2p document sync Users bae123

sync multiple documents:  
  defradb client p2p document sync Users bae123 bae456
```

### Options

```
  -h, --help               help for sync
      --timeout duration   Timeout for sync operations
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client p2p document](defradb_client_p2p_document.md)	 - Configure the P2P document system

