## defradb client blocks get

Get a block by its CID from the blockstore.

```
defradb client blocks get [CID] [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
      --logcolor           Enable colored output
      --logformat string   Log format to use. Options are text, json (default "csv")
      --loglevel string    Log level to use. Options are debug, info, error, fatal (default "info")
      --logoutput string   Log output path (default "stderr")
      --logtrace           Include stacktrace in error and fatal logs
      --rootdir string     Directory for data and configuration to use (default "$HOME/.defradb")
      --url string         URL of the target database's HTTP endpoint (default "localhost:9181")
```

### SEE ALSO

* [defradb client blocks](defradb_client_blocks.md)	 - Interact with the database's blockstore

