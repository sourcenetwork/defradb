# client

Interact with a running DefraDB node as a client

## Synopsis

Interact with a running DefraDB node as a client.
Execute queries, add schema types, and run debug routines.

## Options

```
  -h, --help   help for client
```

## Options inherited from parent commands

```
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --rootdir string       Directory for data and configuration to use (default "$HOME/.defradb")
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

## SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database
* [defradb client blocks](defradb_client_blocks.md)	 - Interact with the database's blockstore
* [defradb client dump](defradb_client_dump.md)	 - Dump the contents of a database node-side
* [defradb client peerid](defradb_client_peerid.md)	 - Get the peer ID of the DefraDB node
* [defradb client ping](defradb_client_ping.md)	 - Ping to test connection to a node
* [defradb client query](defradb_client_query.md)	 - Send a DefraDB GraphQL query request
* [defradb client rpc](defradb_client_rpc.md)	 - Interact with a DefraDB gRPC server
* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a running DefraDB instance

