## defradb client view refresh

Refresh views.

### Synopsis

Refresh views, executing the underlying query and LensVm transforms and
persisting the results.

View is refreshed as the current user, meaning the cached items will reflect that user's
permissions. Subsequent query requests to the view, regardless of user, will receive
items from that cache.

Example: refresh all views
  defradb client view refresh

Example: refresh views by name
  defradb client view refresh --name UserView

Example: refresh views by schema root id
  defradb client view refresh --schema bae123

Example: refresh views by version id. This will also return inactive views
  defradb client view refresh --version bae123
		

```
defradb client view refresh [flags]
```

### Options

```
      --get-inactive     Get inactive views as well as active
  -h, --help             help for refresh
      --name string      View name
      --schema string    View schema Root
      --version string   View version ID
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
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

