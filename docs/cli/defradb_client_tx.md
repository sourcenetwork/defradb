## defradb client tx

Create, commit, and discard DefraDB transactions

### Synopsis

Create, commit, and discard DefraDB transactions

### Options

```
  -h, --help   help for tx
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
      --tx uint              Transaction ID
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client tx commit](defradb_client_tx_commit.md)	 - Commit a DefraDB transaction.
* [defradb client tx create](defradb_client_tx_create.md)	 - Create a new DefraDB transaction.
* [defradb client tx discard](defradb_client_tx_discard.md)	 - Discard a DefraDB transaction.

