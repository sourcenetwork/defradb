## defradb client index list

Shows the list indexes in the database or for a specific collection

### Synopsis

Shows the list indexes in the database or for a specific collection.
		
If the --collection flag is provided, only the indexes for that collection will be shown.
Otherwise, all indexes in the database will be shown.

Example: show all index for 'Users' collection:
  defradb client index list --collection Users

```
defradb client index list [-c --collection <collection>] [flags]
```

### Options

```
  -c, --collection string   Collection name
  -h, --help                help for list
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

