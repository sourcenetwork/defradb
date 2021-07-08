module github.com/sourcenetwork/defradb

go 1.12

require (
	github.com/SierraSoftworks/connor v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/dcrec/secp256k1/v3 v3.0.0
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/go-chi/chi v1.5.2
	github.com/gogo/protobuf v1.2.1
	github.com/graphql-go/graphql v0.7.9
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210324213044-074644c18933
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-blockservice v0.1.0
	github.com/ipfs/go-cid v0.0.3
	github.com/ipfs/go-datastore v0.4.4
	github.com/ipfs/go-ds-badger v0.2.1
	github.com/ipfs/go-ipfs-blockstore v0.0.1
	github.com/ipfs/go-ipfs-ds-help v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipld-format v0.0.2
	github.com/ipfs/go-log v0.0.1
	github.com/ipfs/go-log/v2 v2.1.1
	github.com/ipfs/go-merkledag v0.2.3
	github.com/jbenet/goprocess v0.1.4
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lestrrat-go/jwx v1.2.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multibase v0.0.1
	github.com/multiformats/go-multihash v0.0.13
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go/codec v1.1.7
	github.com/whyrusleeping/go-logging v0.0.0-20170515211332-0457bb6b88fc
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/SierraSoftworks/connor => github.com/sourcenetwork/connor v1.0.3-0.20210312091030-4823d0411a12

	// temp bug fixing
	github.com/graphql-go/graphql => github.com/sourcenetwork/graphql v0.7.10-0.20210312090624-3aa34ef0f75a
// github.com/graphql-go/graphql => github.com/sourcenetwork/graphql v0.7.10-0.20210211004004-07fce0d1409f
)
