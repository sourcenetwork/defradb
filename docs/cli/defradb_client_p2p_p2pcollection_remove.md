## defradb client p2p p2pcollection remove

Remove P2P collections

### Synopsis

Remove P2P collections from the followed pubsub topics.
The removed collections will no longer be synchronized between nodes.

```
defradb client p2p p2pcollection remove [collectionID] [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --rootdir string       Directory for data and configuration to use (default: $HOME/.defradb)
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client p2p p2pcollection](defradb_client_p2p_p2pcollection.md)	 - Configure the P2P collection system

