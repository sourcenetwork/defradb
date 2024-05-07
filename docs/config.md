# DefraDB configuration (YAML)

The default DefraDB directory is `$HOME/.defradb`. It can be changed via the --rootdir CLI flag.

Relative paths are interpreted as being rooted in the DefraDB directory.

## `datastore.store`

Store can be badger or memory. Defaults to `badger`.

- badger: fast pure Go key-value store optimized for SSDs (https://github.com/dgraph-io/badger)
- memory: in-memory version of badger

## `datastore.maxtxnretries`

The number of retries to make in the event of a transaction conflict. Defaults to `5`.

Currently this is only used within the P2P system and will not affect operations initiated by users.

## `datastore.badger.path`

The path to the database data file(s). Defaults to `data`.

## `datastore.badger.valuelogfilesize`

Maximum file size of the value log files.

## `api.address`

Address of the HTTP API to listen on or connect to. Defaults to `127.0.0.1:9181`.

## `api.allowed-origins`

The list of origins a cross-domain request can be executed from.

## `api.pubkeypath`

The path to the public key file for TLS / HTTPS.

## `api.privkeypath`

The path to the private key file for TLS / HTTPS.

## `net.p2pdisabled`

Whether P2P networking is disabled. Defaults to `false`.

## `net.p2paddresses`

List of addresses for the P2P network to listen on. Defaults to `/ip4/127.0.0.1/tcp/9171`.

## `net.pubsubenabled`

Whether PubSub is enabled. Defaults to `true`.

## `net.peers`

List of peers to boostrap with, specified as multiaddresses.

https://docs.libp2p.io/concepts/addressing/

## `net.relay`

Enable libp2p's Circuit relay transport protocol. Defaults to `false`.

https://docs.libp2p.io/concepts/circuit-relay/

## `log.level`

Log level to use. Options are `info` or `error`. Defaults to `info`.

## `log.output`

Log output path. Options are `stderr` or `stdout`. Defaults to `stderr`.

## `log.format`

Log format to use. Options are `text` or `json`. Defaults to `text`.

## `log.stacktrace`

Include stacktrace in error and fatal logs. Defaults to `false`.

## `log.source`

Include source location in logs. Defaults to `false`.

## `log.overrides`

Logger config overrides. Format `<name>,<key>=<val>,...;<name>,...`.

## `log.nocolor`

Disable colored log output. Defaults to `false`.

## `keyring.path`

Path to store encrypted key files in. Defaults to `keys`.

## `keyring.disabled`

Disable the keyring and generate ephemeral keys instead. Defaults to `false`.

## `keyring.namespace`

The service name to use when using the system keyring. Defaults to `defradb`.

## `keyring.backend`

Keyring backend to use. Defaults to `file`.

- `file` Stores keys in encrypted files
- `system` Stores keys in the OS managed keyring
