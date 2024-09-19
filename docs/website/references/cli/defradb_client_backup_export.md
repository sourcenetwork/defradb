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
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --keyring-backend string      Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string    Service name to use when using the system backend (default "defradb")
      --keyring-path string         Path to store encrypted keys when using the file backend (default "keys")
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client backup](defradb_client_backup.md)	 - Interact with the backup utility

