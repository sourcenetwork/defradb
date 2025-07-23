## defradb client acp aac relationship delete

Delete relationship

### Synopsis

Delete relationship

To revoke admin access from a user, we must delete the admin relationship for the target user.
In order to delete the relationship we require all of the following:
1) Relation Name: The type of relation (name must be defined within the admin policy).
2) Target Identity: The identity of the actor the relationship is being deleted with.
3) Requesting Identity: The identity of the actor that is making the request.

Notes:
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.

Example: Revoke admin access from a user:
  defradb client acp relationship delete \
	--relation reader \
	--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
	--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac


```
defradb client acp aac relationship delete [-r --relation] [-a --actor] [-i --identity] [flags]
```

### Options

```
  -a, --actor string      Actor to delete relationship for
  -h, --help              help for delete
  -r, --relation string   Relation that needs to be deleted within the relationship
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

* [defradb client acp aac relationship](defradb_client_acp_aac_relationship.md)	 - Interact with the admin acp relationship features of DefraDB instance

