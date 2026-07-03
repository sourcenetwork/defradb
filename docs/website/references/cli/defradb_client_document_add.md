## defradb client document add

Add a new document.

### Synopsis

Add a new document.

Options:
	-i, --identity
		Marks the document as private and set the identity as the owner. The access to the document
		and permissions are controlled by ACP (Access Control Policy).

	-e, --encrypt
		Encrypt flag specified if the document needs to be encrypted. If set, DefraDB will generate a
		symmetric key for encryption using AES-GCM.

	--encrypt-fields
		Comma-separated list of fields to encrypt. If set, DefraDB will encrypt only the specified fields
		and for every field in the list it will generate a symmetric key for encryption using AES-GCM.
		If combined with '--encrypt' flag, all the fields in the document not listed in '--encrypt-fields'
		will be encrypted with the same key.
		

```
defradb client document add [<document>] [flags]
```

### Examples

```
Add from string1:  
  defradb client document add --collection-name User '{ "name": "Bob" }'

Add from string, with identity:  
  defradb client document add --collection-name User '{ "name": "Bob" }' \
  	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Add multiple from string:  
  defradb client document add --collection-name User '[{ "name": "Alice" }, { "name": "Bob" }]'

Add from file:  
  defradb client document add --collection-name User -f document.json

Add from stdin:  
  cat document.json | defradb client document add --collection-name User -
```

### Options

```
      --collection-id string     Collection ID
      --collection-name string   Collection name
      --enable-signing           Override signing for this operation
  -e, --encrypt                  Flag to enable encryption of the document
      --encrypt-fields strings   Comma-separated list of fields to encrypt
  -f, --file string              File containing document(s)
      --get-inactive             Get inactive collections as well as active
  -h, --help                     help for add
      --version-id string        Collection version ID
```

### Options inherited from parent commands

```
      --audience string             Audience to set on minted auth tokens. Defaults to the host of --url
  -i, --identity string             Hex formatted private key used to authenticate with ACP
      --log-format string           Log format to use. Options are text or json (default "text")
      --log-level string            Log level to use. Options are debug, info, error, fatal (default "info")
      --log-output string           Log output path. Options are stderr or stdout. (default "stderr")
      --log-overrides string        Logger config overrides. Format <name>,<key>=<val>,...;<name>,...
      --log-source                  Include source location in logs
      --log-stacktrace              Include stacktrace in error and fatal logs
      --no-log-color                Disable colored log output
      --rootdir string              Directory for persistent data (default: $HOME/.defradb)
      --source-hub-address string   The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor
      --tx uint                     Transaction ID
      --url string                  URL of HTTP endpoint to listen on or connect to (default "127.0.0.1:9181")
```

### SEE ALSO

* [defradb client document](defradb_client_document.md)	 - Interact with documents.

