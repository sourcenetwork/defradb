## defradb client acp document relationship

Interact with the document acp relationship features of DefraDB instance

### Synopsis

Interact with the document acp relationship features of DefraDB instance

### Options

```
  -h, --help   help for relationship
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

* [defradb client acp document](defradb_client_acp_document.md)	 - Interact with the document access control system of a DefraDB node
* [defradb client acp document relationship add](defradb_client_acp_document_relationship_add.md)	 - Add new relationship
* [defradb client acp document relationship delete](defradb_client_acp_document_relationship_delete.md)	 - Delete relationship

