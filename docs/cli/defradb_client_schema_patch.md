## defradb client schema patch

Patch an existing schema type

### Synopsis

Patch an existing schema.

Uses JSON PATCH formatting as a DDL.

Example: patch from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: patch from file:
  defradb client schema patch -f patch.json

Example: patch from stdin:
  cat patch.json | defradb client schema patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.

```
defradb client schema patch [schema] [flags]
```

### Options

```
  -f, --file string   File to load a patch from
  -h, --help          help for patch
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

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a running DefraDB instance

