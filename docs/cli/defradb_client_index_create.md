## defradb client index create

Creates a secondary index on a collection's field(s)

### Synopsis

Creates a secondary index on a collection's field(s).
		
The --name flag is optional. If not provided, a name will be generated automatically.

Example: create an index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name

Example: create a named index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name --name UsersByName

```
defradb client index create -c --collection <collection> --fields <fields> [-n --name <name>] [flags]
```

### Options

```
  -c, --collection string   Collection name
      --fields string       Fields to index
  -h, --help                help for create
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

