# start

Start a DefraDB node

## Synopsis

Start a new instance of DefraDB node.

```
defradb start [flags]
```

## Options

```
      --email string                Email address used by the CA for notifications (default "example@example.com")
  -h, --help                        help for start
      --max-txn-retries int         Specify the maximum number of retries per transaction (default 5)
      --no-p2p                      Disable the peer-to-peer network synchronization system
      --p2paddr string              Listener address for the p2p network (formatted as a libp2p MultiAddr) (default "/ip4/0.0.0.0/tcp/9171")
      --peers string                List of peers to connect to
      --privkeypath string          Path to the private key for tls (default "certs/server.crt")
      --pubkeypath string           Path to the public key for tls (default "certs/server.key")
      --store string                Specify the datastore to use (supported: badger, memory) (default "badger")
      --tcpaddr string              Listener address for the tcp gRPC server (formatted as a libp2p MultiAddr) (default "/ip4/0.0.0.0/tcp/9161")
      --tls                         Enable serving the API over https
      --valuelogfilesize ByteSize   Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize (default 1GiB)
```

## Options inherited from parent commands

```
      --logformat string     Log format to use. Options are csv, json (default "csv")
      --logger stringArray   Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...
      --loglevel string      Log level to use. Options are debug, info, error, fatal (default "info")
      --lognocolor           Disable colored log output
      --logoutput string     Log output path (default "stderr")
      --logtrace             Include stacktrace in error and fatal logs
      --rootdir string       Directory for data and configuration to use (default "$HOME/.defradb")
      --url string           URL of HTTP endpoint to listen on or connect to (default "localhost:9181")
```

## SEE ALSO

* [defradb](defradb.md)	 - DefraDB Edge Database

