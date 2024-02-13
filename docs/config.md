# DefraDB configuration (YAML)

The default DefraDB directory is `$HOME/.defradb`. It can be changed via the --rootdir CLI flag.

Relative paths are interpreted as being rooted in the DefraDB directory.

## `datastore.store`

Store can be badger or memory. Defaults to `badger`.

- badger: fast pure Go key-value store optimized for SSDs (https://github.com/dgraph-io/badger)
- memory: in-memory version of badger

## `datastore.maxtxnretries`

Maximum number of times to retry a failed transaction. Defaults to `5`.

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