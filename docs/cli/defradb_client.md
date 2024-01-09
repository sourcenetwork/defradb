## defradb client

Interact with a DefraDB node

### Synopsis

Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.

### Options

```
  -h, --help      help for client
      --tx uint   Transaction ID
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

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility
* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.
* [defradb client dump](defradb_client_dump.md)	 - Dump the contents of DefraDB node-side
* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance
* [defradb client p2p](defradb_client_p2p.md)	 - Interact with the DefraDB P2P system
* [defradb client query](defradb_client_query.md)	 - Send a DefraDB GraphQL query request
* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node
* [defradb client tx](defradb_client_tx.md)	 - Create, commit, and discard DefraDB transactions
* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

