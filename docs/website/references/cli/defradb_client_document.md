## defradb client document

Interact with documents.

### Synopsis

Add, read, update, and delete documents within a collection.

### Options

```
  -h, --help              help for document
  -i, --identity string   Hex formatted private key used to authenticate with ACP
      --tx uint           Transaction ID
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
      --keyring-backend string      Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string    Service name to use when using the system backend (default "defradb")
      --keyring-path string         Path (relative to DefraDB root directory) to store encrypted keys when using the file backend (default "keys")
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
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client document add](defradb_client_document_add.md)	 - Add a new document.
* [defradb client document delete](defradb_client_document_delete.md)	 - Delete documents by docID or filter.
* [defradb client document get](defradb_client_document_get.md)	 - View document fields.
* [defradb client document update](defradb_client_document_update.md)	 - Update documents by docID or filter.

