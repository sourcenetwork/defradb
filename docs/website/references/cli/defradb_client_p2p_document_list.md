## defradb client p2p document list

List P2P documents

### Synopsis

List all P2P documents in the pubsub topics.
This is the list of documents of the node that are synchronized on the pubsub network.

```
defradb client p2p document list [flags]
```

### Options

```
  -h, --help   help for list
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
      --remote-dac-address string   Vera address authorized to make Remote DAC transactions on behalf of the actor
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client p2p document](defradb_client_p2p_document.md)	 - Configure the P2P document system

