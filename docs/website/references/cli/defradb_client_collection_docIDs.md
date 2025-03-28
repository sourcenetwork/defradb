## defradb client collection docIDs

List all document IDs (docIDs).

### Synopsis

List all document IDs (docIDs).
		
Example: list all docID(s):
  defradb client collection docIDs --name User

Example: list all docID(s), with an identity:
  defradb client collection docIDs -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f --name User 
		

```
defradb client collection docIDs [-i --identity] [flags]
```

### Options

```
  -h, --help   help for docIDs
```

### Options inherited from parent commands

```
      --get-inactive                Get inactive collections as well as active
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
      --name string                 Collection name
      --no-keyring                  Disable the keyring and generate ephemeral keys
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --schema string               Collection schema Root
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version string              Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

