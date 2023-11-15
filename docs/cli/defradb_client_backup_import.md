## defradb client backup import

Import a JSON data file to the database

### Synopsis

Import a JSON data file to the database.

Example: import data to the database:
  defradb client import user_data.json

```
defradb client backup import <input_path> [flags]
```

### Options

```
  -h, --help   help for import
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

* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility

