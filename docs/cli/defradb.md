## defradb

DefraDB Edge Database

### Synopsis

DefraDB is the edge database to power the user-centric future.

Start a database node, query a local or remote node, and much more.

DefraDB is released under the BSL license, (c) 2022 Democratized Data Foundation.
See https://docs.source.network/BSLv0.2.txt for more information.


### Options

```
  -h, --help               help for defradb
      --logcolor           Enable colored output
      --logformat string   Log format to use. Options are text, json (default "csv")
      --loglevel string    Log level to use. Options are debug, info, error, fatal (default "info")
      --logoutput string   Log output path (default "stderr")
      --logtrace           Include stacktrace in error and fatal logs
      --rootdir string     Directory for data and configuration to use (default "$HOME/.defradb")
      --url string         URL of the target database's HTTP endpoint (default "localhost:9181")
```

### SEE ALSO

* [defradb client](defradb_client.md)	 - Interact with a running DefraDB node as a client
* [defradb init](defradb_init.md)	 - Initialize DefraDB's root directory and configuration file
* [defradb server-dump](defradb_server-dump.md)	 - Dumps the state of the entire database
* [defradb start](defradb_start.md)	 - Start a DefraDB node
* [defradb version](defradb_version.md)	 - Display the version information of DefraDB and its components

