## defradb client

Interact with a DefraDB node

### Synopsis

Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.

### Options

```
  -h, --help   help for client
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
* [defradb client blocks](defradb_client_blocks.md)	 - Interact with the database's blockstore
* [defradb client dump](defradb_client_dump.md)	 - Dump the contents of DefraDB node-side
* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance
* [defradb client peerid](defradb_client_peerid.md)	 - Get the PeerID of the node
* [defradb client ping](defradb_client_ping.md)	 - Ping to test connection with a node
* [defradb client query](defradb_client_query.md)	 - Send a DefraDB GraphQL query request
* [defradb client rpc](defradb_client_rpc.md)	 - Interact with a DefraDB node via RPC
* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node

