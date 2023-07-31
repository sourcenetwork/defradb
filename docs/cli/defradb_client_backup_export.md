## defradb client backup export

Export the database to a file

### Synopsis

Export the database to a file. If a file exists at the <output_path> location, it will be overwritten.
		
If the --collection flag is provided, only the data for that collection will be exported.
Otherwise, all collections in the database will be exported.

If the --pretty flag is provided, the JSON will be pretty printed.

Example: export data for the 'Users' collection:
  defradb client export --collection Users user_data.json

```
defradb client backup export  [-c --collections | -p --pretty | -f --format] <output_path> [flags]
```

### Options

```
  -c, --collections strings   List of collections
  -f, --format string         Define the output format. Supported formats: [json] (default "json")
  -h, --help                  help for export
  -p, --pretty                Set the output JSON to be pretty printed
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

* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility

