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

## `log.colordisabled`

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

## `lens.runtime`

The LensVM wasm runtime to run lens modules in.

Possible values:
- `wasm-time` (default): https://github.com/bytecodealliance/wasmtime-go
- `wasmer` (windows not supported): https://github.com/wasmerio/wasmer-go
- `wazero`: https://github.com/tetratelabs/wazero

## `acp.type`

The type of ACP module to use.

Possible values:
- `none` (default): No ACP
- `local` local-only ACP
- `source-hub` source hub ACP: https://github.com/sourcenetwork/sourcehub

## `acp.sourceHub.ChainID`

The ID of the SourceHub chain to store ACP data in. Required when using `acp.type`:`source-hub`.

## `acp.sourceHub.GRPCAddress`

The address of the SourceHub GRPC server. Required when using `acp.type`:`source-hub`.

## `acp.sourceHub.CometRPCAddress`

The address of the SourceHub Comet RPC server. Required when using `acp.type`:`source-hub`.

## `acp.sourceHub.KeyName`

The name of the key in the keyring where the SourceHub credentials used to sign (and pay for) SourceHub
transactions created by the node is stored. Required when using `acp.type`:`source-hub`.

## `acp.sourceHub.address`

The SourceHub address of the actor that client-side actions should permit to make SourceHub actions on
their behalf.  This is a client-side only config param.  It is required if the client wishes to make
SourceHub ACP requests in order to create protected data.
