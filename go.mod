module github.com/sourcenetwork/defradb

go 1.12

require (
	github.com/SierraSoftworks/connor v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/go-chi/chi v1.5.2
	github.com/graphql-go/graphql v0.7.9
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.5.0
	github.com/ipfs/go-ipfs-blockstore v1.1.0
	github.com/ipfs/go-ipfs-ds-help v1.1.0
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log v1.0.5
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/ipfs/go-merkledag v0.5.0
	github.com/jbenet/goprocess v0.1.4
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multibase v0.0.3
	github.com/multiformats/go-multihash v0.0.15
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go/codec v1.1.7
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/SierraSoftworks/connor => github.com/sourcenetwork/connor v1.0.3-0.20210312091030-4823d0411a12

	// temp bug fixing
	github.com/graphql-go/graphql => github.com/sourcenetwork/graphql v0.7.10-0.20220122211559-2fe60b2360cc
// github.com/graphql-go/graphql => github.com/sourcenetwork/graphql v0.7.10-0.20210211004004-07fce0d1409f
)
