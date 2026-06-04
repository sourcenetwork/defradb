## defradb client collection delete

Delete collections

### Synopsis

Delete one or more collections by name.

A single name, or a comma-separated list of names, may be provided. All named
collections are removed atomically in a single operation. This can be used to
delete collections that reference each other via relations, since deleting them
one at a time would leave a dangling reference and be rolled back.

By default, every version of each named collection is deleted (active head and
all earlier versions). Pass --active-only to delete only the latest (head) version
and keep earlier versions intact.

The named collections must not contain any documents. Delete all documents first
before deleting the collection.

```
defradb client collection delete [collectionNames] [flags]
```

### Examples

```
delete every version of a single collection:  
  defradb client collection delete Users

delete every version of multiple collections in one call (this can be used to delete collections that reference each other via relations):  
  defradb client collection delete Users,Books

delete only the active head version, keeping earlier versions:  
  defradb client collection delete --active-only Users
```

### Options

```
      --active-only   Delete only the active head version of each named collection (default deletes every version)
  -h, --help          help for delete
```

### Options inherited from parent commands

```
      --collection-id string        Collection ID
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
      --secret-file string          Path to the file containing secrets (default ".env")
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version-id string           Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

