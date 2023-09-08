## defradb client backup

Interact with the backup utility

### Synopsis

Export to or Import from a backup file.
Currently only supports JSON format.

### Options

```
  -h, --help   help for backup
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

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client backup export](defradb_client_backup_export.md)	 - Export the database to a file
* [defradb client backup import](defradb_client_backup_import.md)	 - Import a JSON data file to the database

