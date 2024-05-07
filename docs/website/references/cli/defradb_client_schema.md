# client schema

Interact with the schema system of a running DefraDB instance

## Synopsis

Make changes, updates, or look for existing schema types to a DefraDB node.

## Options

```
  -h, --help   help for schema
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

* [defradb client](defradb_client.md)	 - Interact with a running DefraDB node as a client
* [defradb client schema add](defradb_client_schema_add.md)	 - Add a new schema type to DefraDB
* [defradb client schema patch](defradb_client_schema_patch.md)	 - Patch an existing schema type

