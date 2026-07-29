## defradb client backup export

Export the database to a file

### Synopsis

Export the database to a file.

The backup captures a snapshot of documents in the database, but does not include their
history or ownership information, and docIDs may not be preserved.

The instance must be running in development mode for a backup to be exported.

If a file exists at the `<output_path>` location, it will be overwritten.

If the --collection flag is provided, only the data for that collection will be exported.
Otherwise, all collections in the database will be exported.

If the --pretty flag is provided, the JSON will be pretty printed.


```
defradb client backup export <output_path> [flags]
```

### Examples

```
Export data for the 'Users' collection:  
  defradb client backup export --collections Users user_data.json
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
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility

