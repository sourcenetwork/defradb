# client schema add

Add a new schema type to DefraDB

## Synopsis

Add a new schema type to DefraDB.

Example: add from an argument string:
  defradb client schema add 'type Foo { ... }'

Example: add from file:
  defradb client schema add -f schema.graphql

Example: add from stdin:
  cat schema.graphql | defradb client schema add -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.

```
defradb client schema add [schema] [flags]
```

## Options

```
  -f, --file string   File to load a schema from
  -h, --help          help for add
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

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a running DefraDB instance

