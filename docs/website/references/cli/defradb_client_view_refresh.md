## defradb client view refresh

Refresh views.

### Synopsis

Refresh views, executing the underlying query and LensVm transforms and
persisting the results.

View is refreshed as the current user, meaning the cached items will reflect that user's
permissions. Subsequent query requests to the view, regardless of user, will receive
items from that cache.

```
defradb client view refresh [flags]
```

### Examples

```
refresh all views:  
  defradb client view refresh

refresh views by name:  
  defradb client view refresh --collection-name UserView

refresh views by collection id:  
  defradb client view refresh --collection-id bae123

refresh views by version id:  
  defradb client view refresh --version-id bae123
```

### Options

```
      --collection-id string     Collection ID
      --collection-name string   Collection name
      --get-inactive             Get inactive collections as well as active
  -h, --help                     help for refresh
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

* [defradb client view](defradb_client_view.md)	 - Manage views within a running DefraDB instance

