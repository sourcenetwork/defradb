## defradb client index new

Make a new index on a collection's field(s)

### Synopsis

Make a new index on a collection's field(s).

The --name flag is optional. If not provided, a name will be generated automatically.
The --unique flag is optional. If provided, the index will be unique.
If no order is specified for the field, the default value will be "ASC"

The --vector flag makes a vector index (on a single field, never unique). Its value is the index
config as JSON. Give the config for the algorithm you want under its own key; HNSW is the only one
today, e.g. '{"Metric":"COSINE","Dimensions":3,"HNSW":{}}'. Metric and Dimensions are the essentials
(Metric is the distance metric, COSINE today; Dimensions is the vector length). The HNSW tuning
params are optional and default if omitted:
  M               links per node (higher = better recall, more memory and slower build); default 16
  EfConstruction  build-time search width (higher = better graph, slower build); default 128
  EfSearch        query-time search width (higher = better recall, slower queries); default 64

The --full-text flag makes a ranked full-text index on one String field. Its value is JSON; an
empty object selects BM25 defaults, while parameters may be set with
'{"Algorithm":"BM25","BM25":{"K1":1.2,"B":0.75}}'.

The --trigram flag makes a trigram candidate index on one String field for _like, _ilike, and
_regex filters. Vector, full-text, and trigram modes are mutually exclusive.

The index is built in the background. This command returns once the index is recorded, before
existing documents are indexed. The index starts "building" and becomes "ready" once complete, or
"failed" if it cannot be built. Use 'index list' to check its status.

```
defradb client index new [flags]
```

### Examples

```
make a new index for 'Users' collection on 'name' field:  
  defradb client index new --collection Users --fields name

make a new named index for 'Users' collection on 'name' field:  
  defradb client index new --collection Users --fields name --name UsersByName

make a new unique index for 'Users' collection on 'name' and 'age':  
  defradb client index new --collection Users --fields name:ASC,age:DESC --unique

make a new vector index for 'Users' collection on 'vec' field (HNSW defaults):  
  defradb client index new --collection Users --fields vec --vector '{"Metric":"COSINE","Dimensions":3,"HNSW":{}}'

make a new vector index tuning the HNSW params:  
  defradb client index new --collection Users --fields vec --vector '{"Metric":"COSINE","Dimensions":3,"HNSW":{"M":16,"EfConstruction":128,"EfSearch":64}}'

make a BM25 full-text index for 'Articles' on 'body':  
  defradb client index new --collection Articles --fields body --full-text '{}'

make a trigram pattern index for 'Users' on 'name':  
  defradb client index new --collection Users --fields name --trigram
```

### Options

```
  -c, --collection string   Collection name
      --fields strings      Fields to index
      --full-text string    Full-text index config as JSON (makes a ranked index)
  -h, --help                help for new
  -n, --name string         Index name
      --trigram             Make a trigram pattern-matching index
  -u, --unique              Make the index unique
      --vector string       Vector index config as JSON (makes a vector index)
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

* [defradb client index](defradb_client_index.md)	 - Manage collections' indexes of a running DefraDB instance

