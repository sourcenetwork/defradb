## defradb client schema describe

View schema descriptions.

### Synopsis

Introspect schema types.

Example: view all schemas
  defradb client schema describe
		
Example: view schemas by name
  defradb client schema describe --name User
		
Example: view schemas by root
  defradb client schema describe --root bae123
		
Example: view a single schema by version id
  defradb client schema describe --version bae123
		

```
defradb client schema describe [flags]
```

### Options

```
  -h, --help             help for describe
      --name string      Schema name
      --root string      Schema root
      --version string   Schema Version ID
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
      --tx uint              Transaction ID
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

### SEE ALSO

* [defradb client schema](defradb_client_schema.md)	 - Interact with the schema system of a DefraDB node

