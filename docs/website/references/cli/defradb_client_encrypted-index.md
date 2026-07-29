## defradb client encrypted-index

Manage collections' encrypted indexes on a running DefraDB node

### Synopsis

Manage collections' encrypted indexes on a running DefraDB node.

### Options

```
  -h, --help   help for encrypted-index
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

* [defradb client](defradb_client.md)	 - Interact with a DefraDB node
* [defradb client encrypted-index delete](defradb_client_encrypted-index_delete.md)	 - Delete an encrypted index from a collection's field
* [defradb client encrypted-index list](defradb_client_encrypted-index_list.md)	 - Lists the encrypted indexes in the database or for a specific collection
* [defradb client encrypted-index new](defradb_client_encrypted-index_new.md)	 - Make a new encrypted index on a collection's field

