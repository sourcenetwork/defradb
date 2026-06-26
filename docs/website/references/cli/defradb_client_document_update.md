## defradb client document update

Update documents by docID or filter.

### Synopsis

Update documents by docID or filter.

```
defradb client document update [flags]
```

### Examples

```
update by filter:  
  defradb client document update --collection-name User \
  --filter '{ "points": { "_gte": 100 } }' --updater '{ "verified": true }'

update by docID:  
  defradb client document update --collection-name User \
  --docID bae-123 --updater '{ "verified": true }'

update private docID, with identity:  
  defradb client document update --collection-name User \
  -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f \
  --docID bae-123 --updater '{ "verified": true }'
```

### Options

```
      --collection-id string     Collection ID
      --collection-name string   Collection name
      --docID string             Document ID
      --filter string            Document filter
      --get-inactive             Get inactive collections as well as active
  -h, --help                     help for update
      --updater string           Document updater
      --version-id string        Collection version ID
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

* [defradb client document](defradb_client_document.md)	 - Interact with documents.

