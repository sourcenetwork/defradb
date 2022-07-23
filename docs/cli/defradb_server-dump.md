## defradb server-dump

Dumps the state of the entire database

```
defradb server-dump [flags]
```

### Options

```
  -h, --help           help for server-dump
      --store string   Datastore to use. Options are badger, memory (default "badger")
```

### Options inherited from parent commands

```
      --logcolor           Enable colored output
      --logformat string   Log format to use. Options are text, json (default "csv")
      --loglevel string    Log level to use. Options are debug, info, error, fatal (default "info")
      --logoutput string   Log output path (default "stderr")
      --logtrace           Include stacktrace in error and fatal logs
      --rootdir string     Directory for data and configuration to use (default "$HOME/.defradb")
      --url string         URL of the target database's HTTP endpoint (default "localhost:9181")
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

