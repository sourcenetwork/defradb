## defradb client index drop

Drop a collection's secondary index

### Synopsis

Drop a collection's secondary index.
		
Example: drop the index 'UsersByName' for 'Users' collection:
  defradb client index create --collection Users --name UsersByName

```
defradb client index drop -c --collection <collection> -n --name <name> [flags]
```

### Options

```
  -c, --collection string   Collection name
  -h, --help                help for drop
  -n, --name string         Index name
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

* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance

