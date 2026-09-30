## defradb client collection truncate

Truncate the given collection

### Synopsis

Truncate the given collection, removing document data from the local node.
Without a filter all documents are removed. With a filter only matching documents and their
unshared history are removed. Changes do not propagate to other nodes.

```
defradb client collection truncate [flags]
```

### Options

```
      --collection-id string     Collection ID
      --collection-name string   Collection name
      --filter string            Document filter
      --get-inactive             Get inactive collections as well as active
  -h, --help                     help for truncate
      --version-id string        Collection version ID
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

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

