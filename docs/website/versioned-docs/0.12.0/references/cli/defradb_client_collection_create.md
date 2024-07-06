## defradb client collection create

Create a new document.

### Synopsis

Create a new document.
		
Options:
    -i, --identity 
        Marks the document as private and set the identity as the owner. The access to the document
		and permissions are controlled by ACP (Access Control Policy).

	-e, --encrypt
		Encrypt flag specified if the document needs to be encrypted. If set, DefraDB will generate a
		symmetric key for encryption using AES-GCM.

Example: create from string:
  defradb client collection create --name User '{ "name": "Bob" }'

Example: create from string, with identity:
  defradb client collection create --name User '{ "name": "Bob" }' \
  	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Example: create multiple from string:
  defradb client collection create --name User '[{ "name": "Alice" }, { "name": "Bob" }]'

Example: create from file:
  defradb client collection create --name User -f document.json

Example: create from stdin:
  cat document.json | defradb client collection create --name User -
		

```
defradb client collection create [-i --identity] [-e --encrypt] <document> [flags]
```

### Options

```
  -e, --encrypt       Flag to enable encryption of the document
  -f, --file string   File containing document(s)
  -h, --help          help for create
```

### Options inherited from parent commands

```
      --get-inactive               Get inactive collections as well as active
  -i, --identity string            Hex formatted private key used to authenticate with ACP
      --keyring-backend string     Keyring backend to use. Options are file or system (default "file")
      --keyring-namespace string   Service name to use when using the system backend (default "defradb")
      --keyring-path string        Path to store encrypted keys when using the file backend (default "keys")
      --log-format string          Log format to use. Options are text or json (default "text")
      --log-level string           Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string          Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string       Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                 Include source location in logs
      --log-stacktrace             Include stacktrace in error and fatal logs
      --name string                Collection name
      --no-keyring                 Disable the keyring and generate ephemeral keys
      --no-log-color               Disable colored log output
      --rootdir string             Directory for persistent data (default: $HOME/.defradb)
      --schema string              Collection schema Root
      --tx uint                    Transaction ID
      --url string                 URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
      --version string             Collection version ID
```

### SEE ALSO

* [defradb client collection](defradb_client_collection.md)	 - Interact with a collection.

