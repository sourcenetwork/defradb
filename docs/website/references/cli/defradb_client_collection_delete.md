## defradb client collection delete

Delete documents by docID or filter.

### Synopsis

Delete documents by docID or filter and lists the number of documents deleted.
		
Example: delete by docID:
  defradb client collection delete  --name User --docID bae-123

Example: delete by docID with identity:
  defradb client collection delete -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j --name User --docID bae-123

Example: delete by filter:
  defradb client collection delete --name User --filter '{ "_gte": { "points": 100 } }'
		

```
defradb client collection delete [-i --identity] [--filter <filter> --docID <docID>] [flags]
```

### Options

```
      --docID string    Document ID
      --filter string   Document filter
  -h, --help            help for delete
```

### Options inherited from parent commands

```
      --get-inactive               Get inactive collections as well as active
  -i, --identity string            ACP Identity
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --name string                Collection name
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --no-log-color               Disable colored log output
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --schema string              Collection schema Root
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version string             Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

