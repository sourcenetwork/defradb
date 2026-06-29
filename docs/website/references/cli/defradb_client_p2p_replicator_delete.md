## defradb client p2p replicator delete

Delete replicator(s) and stop synchronization

### Synopsis

Delete replicator(s) and stop synchronization.
A replicator synchronizes one or all collection(s) from this instance to another.

```
defradb client p2p replicator delete <peerID> [flags]
```

### Examples

```
Delete replicator:  
  defradb client p2p replicator delete -c Users 12D3...
```

### Options

```
  -c, --collection strings   Collection(s) to stop replicating
  -h, --help                 help for delete
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

* [defradb client p2p replicator](defradb_client_p2p_replicator.md)	 - Configure the replicator system

