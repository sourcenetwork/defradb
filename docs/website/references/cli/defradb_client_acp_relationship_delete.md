## defradb client acp relationship delete

Delete relationship

### Synopsis

Delete relationship

To revoke access to a document for an actor, we must delete the relationship between the
actor and the document. In order to delete the relationship we require all of the following:

1) Target DocID: The docID of the document we want to delete a relationship for.
2) Collection Name: The name of the collection that has the Target DocID.
3) Relation Name: The type of relation (name must be defined within the linked policy on collection).
4) Target Identity: The identity of the actor the relationship is being deleted for.
5) Requesting Identity: The identity of the actor that is making the request.

Notes:
  - ACP must be available (i.e. ACP can not be disabled).
  - The target document must be registered with ACP already (policy & resource specified).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the relationship record was not found, then it will be a no-op.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: Let another actor (4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5) read a private document:
  defradb client acp relationship delete \
	--collection Users \
	--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
	--relation reader \
	--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
	--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac


```
defradb client acp relationship delete [--docID] [-c --collection] [-r --relation] [-a --actor] [-i --identity] [flags]
```

### Options

```
  -a, --actor string        Actor to delete relationship for
  -c, --collection string   Collection that has the resource and policy for object
      --docID string        Document Identifier (ObjectID) to delete relationship for
  -h, --help                help for delete
  -r, --relation string     Relation that needs to be deleted within the relationship
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

* [defradb client acp relationship](defradb_client_acp_relationship.md)	 - Interact with the acp relationship features of DefraDB instance

