## defradb server-dump

Dumps the state of the entire database

```
defradb server-dump [flags]
```

### Options

```
  -h, --help   help for server-dump
```

### Options inherited from parent commands

```
      --log-format string      Log format to use. Options are text or json (default "text")
      --log-level string       Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string      Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string   Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source             Include source location in logs
      --log-stacktrace         Include stacktrace in error and fatal logs
      --no-log-color           Disable colored log output
      --rootdir string         Directory for persistent data (default: $HOME/.defradb)
```

### SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

