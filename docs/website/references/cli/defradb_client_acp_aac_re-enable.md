## defradb client acp aac re-enable

Re-enable the admin access control

### Synopsis

Re-enable the admin access control

Example:
  defradb client acp aac re-enable -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Note:
- This command will re-enable an already configured admin acp system that is temporarily disabled.
- If admin acp is already enabled, then it will return an error.
- If admin acp is in a clean/non-configured state, then it will return an error.

Learn more about the DefraDB [ACP System](/acp/README.md)



```
defradb client acp aac re-enable [-i --identity] [flags]
```

### Options

```
  -h, --help   help for re-enable
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

* [defradb client acp aac](defradb_client_acp_aac.md)	 - Interact with the admin access control system of a DefraDB node

