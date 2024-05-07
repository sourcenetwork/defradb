# defradb

DefraDB Edge Database

## Synopsis

DefraDB is the edge database to power the user-centric future.

Start a database node, issue a request to a local or remote node, and much more.

DefraDB is released under the BSL license, (c) 2022 Democratized Data Foundation.
See https://docs.source.network/BSL.txt for more information.


## Options

```
  -h, --help                 help for defradb
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

* [defradb client](defradb_client.md)	 - Interact with a running DefraDB node as a client
* [defradb init](defradb_init.md)	 - Initialize DefraDB's root directory and configuration file
* [defradb server-dump](defradb_server-dump.md)	 - Dumps the state of the entire database
* [defradb start](defradb_start.md)	 - Start a DefraDB node
* [defradb version](defradb_version.md)	 - Display the version information of DefraDB and its components

