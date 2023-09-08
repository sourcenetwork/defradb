## defradb client tx create

Create a new DefraDB transaction.

### Synopsis

Create a new DefraDB transaction.

```
defradb client tx create [flags]
```

### Options

```
      --concurrent   Transaction is concurrent
  -h, --help         help for create
      --read-only    Transaction is read only
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

* [defradb client tx](defradb_client_tx.md)	 - Create, commit, and discard DefraDB transactions

