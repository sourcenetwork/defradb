module github.com/sourcenetwork/defradb

go 1.23.6

require (
	github.com/bits-and-blooms/bitset v1.22.0
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/cyware/ssi-sdk v0.0.0-20231229164914-f93f3006379f
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0
	github.com/dgraph-io/badger/v4 v4.6.0
	github.com/evanphx/json-patch/v5 v5.9.11
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/go-errors/errors v1.5.1
	github.com/gofrs/uuid/v5 v5.3.1
	github.com/iancoleman/strcase v0.3.0
	github.com/ipfs/boxo v0.29.1
	github.com/ipfs/go-block-format v0.2.0
	github.com/ipfs/go-cid v0.5.0
	github.com/ipfs/go-datastore v0.8.2
	github.com/ipfs/go-ipld-format v0.6.0
	github.com/ipld/go-ipld-prime v0.21.0
	github.com/ipld/go-ipld-prime/storage/bsadapter v0.0.0-20240322071758-198d7dba8fb8
	github.com/lens-vm/lens/host-go v0.0.0-20240605170614-979fa9eb14d5
	github.com/lestrrat-go/jwx/v2 v2.1.4
	github.com/libp2p/go-libp2p v0.41.0
	github.com/multiformats/go-multiaddr v0.15.0
	github.com/multiformats/go-multibase v0.2.0
	github.com/multiformats/go-multicodec v0.9.0
	github.com/multiformats/go-multihash v0.2.3
	github.com/onsi/gomega v1.36.3
	github.com/philippgille/chromem-go v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/sourcenetwork/corekv v0.1.1
	github.com/sourcenetwork/corelog v0.0.8
	github.com/sourcenetwork/graphql-go v0.7.10-0.20241003221550-224346887b4a
	github.com/sourcenetwork/immutable v0.3.0
	github.com/stretchr/testify v1.10.0
	github.com/valyala/fastjson v1.6.4
	golang.org/x/crypto v0.36.0
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgraph-io/ristretto/v2 v2.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.6 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.15.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190812055157-5d271430af9f // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hyperledger/aries-framework-go v0.3.2 // indirect
	github.com/hyperledger/aries-framework-go/component/kmscrypto v0.0.0-20230427134832-0c9969493bd3 // indirect
	github.com/hyperledger/aries-framework-go/component/log v0.0.0-20230427134832-0c9969493bd3 // indirect
	github.com/hyperledger/aries-framework-go/component/models v0.0.0-20230501135648-a9a7ad029347 // indirect
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20230427134832-0c9969493bd3 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/go-ipfs-util v0.0.3 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipfs/go-metrics-interface v0.3.0 // indirect
	github.com/jorrizza/ed2curve25519 v0.1.0 // indirect
	github.com/kilic/bls12-381 v0.1.1-0.20210503002446-7b7597926c69 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/lmittmann/tint v1.0.4 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multistream v0.6.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/piprate/json-gold v0.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.4.0 // indirect
)
