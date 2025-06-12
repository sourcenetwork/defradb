// NOTE: This was copied from github.com/cyware/ssi-sdk which is no longer available.
// The usage of the code is this file, in it's current form, is only temporary until we
// find an more appropriate solution.

package crypto

import (
	"fmt"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-varint"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	// Multicodec reference https://github.com/multiformats/multicodec/blob/master/table.csv
	Ed25519MultiCodec   = multicodec.Ed25519Pub
	X25519MultiCodec    = multicodec.X25519Pub
	SECP256k1MultiCodec = multicodec.Secp256k1Pub
	P256MultiCodec      = multicodec.P256Pub
	P384MultiCodec      = multicodec.P384Pub
	P521MultiCodec      = multicodec.P521Pub
	RSAMultiCodec       = multicodec.RsaPub
	SHA256MultiCodec    = multicodec.Sha2_256
)

const (
	Ed25519        KeyType = "Ed25519"
	X25519         KeyType = "X25519"
	SECP256k1      KeyType = "secp256k1"
	SECP256k1ECDSA KeyType = "secp256k1-ECDSA"
	P224           KeyType = "P-224"
	P256           KeyType = "P-256"
	P384           KeyType = "P-384"
	P521           KeyType = "P-521"
	RSA            KeyType = "RSA"
	BLS12381G1     KeyType = "BLS12381G1"
	BLS12381G2     KeyType = "BLS12381G2"
	Dilithium2     KeyType = "Dilithium2"
	Dilithium3     KeyType = "Dilithium3"
	Dilithium5     KeyType = "Dilithium5"

	RSAKeySize int = 2048
)

const (
	// Prefix did:key prefix
	Prefix = "did:key"
	// Base58BTCMultiBase Base58BTC https://github.com/multiformats/go-multibase/blob/master/multibase.go
	Base58BTCMultiBase = multibase.Base58BTC
)

type (
	DIDKey string
)

func (d DIDKey) String() string {
	return string(d)
}

// createDIDKey constructs a did:key from a specific key type and its corresponding public key
// This method does not attempt to validate that the provided public key is of the specified key type.
// A safer method is `GenerateDIDKey` which handles key generation based on the provided key type.
func createDIDKey(kt KeyType, publicKey []byte) (*DIDKey, error) {
	if !isSupportedDIDKeyType(kt) {
		return nil, fmt.Errorf("unsupported did:key type: %s", kt)
	}

	// did:key:<multibase encoded, multicodec identified, public key>
	encoded, err := multibaseEncodedKey(kt, publicKey)
	if err != nil {
		return nil, errors.Wrap("multibase encoding key", err)
	}
	didKey := DIDKey(fmt.Sprintf("%s:%s", Prefix, encoded))
	return &didKey, nil
}

func isSupportedDIDKeyType(kt KeyType) bool {
	keyTypes := getSupportedDIDKeyTypes()
	for _, t := range keyTypes {
		if t == kt {
			return true
		}
	}
	return false
}

func getSupportedDIDKeyTypes() []KeyType {
	return []KeyType{Ed25519, X25519, SECP256k1,
		P256, P384, P521, RSA}
}

// MultibaseEncodedKey takes a key type and a public key value and returns the multibase encoded key
func multibaseEncodedKey(kt KeyType, publicKey []byte) (string, error) {
	multiCodec, err := keyTypeToMultiCodec(kt)
	if err != nil {
		return "", fmt.Errorf("could find mutlicodec for key type<%s>", kt)
	}
	prefix := varint.ToUvarint(uint64(multiCodec))
	codec := append(prefix, publicKey...)
	encoded, err := multibase.Encode(Base58BTCMultiBase, codec)
	if err != nil {
		return "", errors.Wrap("multibase encoding", err)
	}
	return encoded, nil
}

func keyTypeToMultiCodec(kt KeyType) (multicodec.Code, error) {
	switch kt {
	case Ed25519:
		return Ed25519MultiCodec, nil
	case X25519:
		return X25519MultiCodec, nil
	case SECP256k1:
		return SECP256k1MultiCodec, nil
	case P256:
		return P256MultiCodec, nil
	case P384:
		return P384MultiCodec, nil
	case P521:
		return P521MultiCodec, nil
	case RSA:
		return RSAMultiCodec, nil
	}
	return 0, fmt.Errorf("unknown multicodec for key type: %s", kt)
}
