## defradb client acp document

Interact with the document access control system of a DefraDB node

### Synopsis

Interact with the document access control system of a DefraDB node

Learn more about the
[Document Access Control](https://docs.source.network/defradb/security/document-access-control/)
system.

		

### Options

```
  -h, --help   help for document
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

* [defradb client acp](defradb_client_acp.md)	 - Interact with the access control system(s) of a DefraDB node
* [defradb client acp document policy](defradb_client_acp_document_policy.md)	 - Interact with the document acp policy features of DefraDB instance
* [defradb client acp document relationship](defradb_client_acp_document_relationship.md)	 - Interact with the document acp relationship features of DefraDB instance

