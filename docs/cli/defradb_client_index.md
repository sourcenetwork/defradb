## defradb client index

Manage collections' indexes of a running DefraDB instance

### Synopsis

Manage (create, drop, or list) collection indexes on a DefraDB node.

### Options

```
  -h, --help   help for index
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

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client index create](defradb_client_index_create.md)	 - Creates a secondary index on a collection's field(s)
* [defradb client index drop](defradb_client_index_drop.md)	 - Drop a collection's secondary index
* [defradb client index list](defradb_client_index_list.md)	 - Shows the list indexes in the database or for a specific collection

