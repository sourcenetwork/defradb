## defradb client view add

Add new view

### Synopsis

Add new database view.

Example: add from an argument string:
  defradb client view add 'Foo { name, ...}' 'type Foo { ... }'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.

```
defradb client view add [query] [sdl] [flags]
```

### Options

```
  -h, --help   help for add
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

* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

