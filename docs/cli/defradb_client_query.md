## defradb client query

Send a DefraDB GraphQL query request

### Synopsis

Send a DefraDB GraphQL query request to the database.

A query request can be sent as a single argument. Example command:
  defradb client query 'query { ... }'

Do a query request from a file by using the '-f' flag. Example command:
  defradb client query -f request.graphql

Or it can be sent via stdin by using the '-' special syntax. Example command:
  cat request.graphql | defradb client query -

A GraphQL client such as GraphiQL (https://github.com/graphql/graphiql) can be used to interact
with the database more conveniently.

To learn more about the DefraDB GraphQL Query Language, refer to https://docs.source.network.

```
defradb client query [query request] [flags]
```

### Options

```
  -f, --file string   File containing the query request
  -h, --help          help for query
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

* [defradb client](defradb_client.md)	 - Interact with a running DefraDB node as a client

