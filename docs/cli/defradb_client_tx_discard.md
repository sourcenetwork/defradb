## defradb client tx discard

Discard a DefraDB transaction.

### Synopsis

Discard a DefraDB transaction.

```
defradb client tx discard [id] [flags]
```

### Options

```
  -h, --help   help for discard
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

