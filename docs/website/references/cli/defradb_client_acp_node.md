## defradb client acp node

Interact with the node access control system of a DefraDB node

### Synopsis

Interact with the node access control system of a DefraDB node
Note:
- Clean state means, if the access control was never configured, or the entire state was purged.
- The check the node acp status use the 'client acp node status' command
- To start node acp for the first time, use the start command with '--acp-enable true'.
- Specifying an identity is a MUST, when starting first time (from a clean state), this identity
will become the node owner identity.
- To temporarily disable node acp, use the 'client acp node disable' command.
- To re-enable node acp when it is temporarily disabled, use the 'client acp node re-enable' command.
- To give node access to other users use the 'client acp node relationship add' command.
- To revoke node access from other users use the 'client acp node relationship delete' command.
- To reset/purge acp state into a clean state, use the 'client acp node purge' command.
- Purge command(s) require the user to be in dev-mode (Warning: all state will be lost).

For quick help: 'defradb client acp node --help'

Learn more about the DefraDB [ACP System](/acp/README.md)

		

### Options

```
  -h, --help   help for node
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

* [defradb client acp](defradb_client_acp.md)	 - Interact with the access control system(s) of a DefraDB node
* [defradb client acp node disable](defradb_client_acp_node_disable.md)	 - Disable the node access control
* [defradb client acp node re-enable](defradb_client_acp_node_re-enable.md)	 - Re-enable the node access control
* [defradb client acp node relationship](defradb_client_acp_node_relationship.md)	 - Interact with the node acp relationship features of DefraDB instance
* [defradb client acp node status](defradb_client_acp_node_status.md)	 - Check the node access control status

