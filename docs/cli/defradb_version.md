## defradb version

Display the version information of DefraDB and its components

```
defradb version [flags]
```

### Options

```
  -f, --format string   Version output format. Options are text, json
      --full            Display the full version information
  -h, --help            help for version
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

