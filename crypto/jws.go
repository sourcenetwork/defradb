package crypto

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/crypto/did"

	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
)

var (
	ErrMissingSignatures                    = errors.New("Provided JWS message has no signatures")
	ErrUnsupportedJWSAlgorithm              = errors.New("The given 'alg' property in the JWS header doesn't match the allowed algorithms")
	ErrUnsupportedDIDVerificationMethodType = errors.New("The 'Type' field on the given verificationMethod object isn't supported")
	ErrMissingDIDVerificationMethod         = errors.New("The given DID Document doesn't contain a matching verification method")
	ErrInvalidKID                           = errors.New("The given kid doesn't contain a valid DID identifer")
	ErrDIDMissingKeyID                      = errors.New("The given DID identifier is missing the #<key> ID suffix")
)

var (
	linkedDataSignaturesToJWA = map[string]jwa.SignatureAlgorithm{
		"EcdsaSecp256k1VerificationKey2019": jwa.ES256K,
		"Ed25519VerificationKey2018":        jwa.EdDSA,
	}
	supportedJWA = map[jwa.SignatureAlgorithm]bool{
		jwa.ES256K: true,
		jwa.EdDSA:  true,
	}
	supportedJWAString = map[string]jwa.SignatureAlgorithm{
		"ES256K": jwa.ES256K,
		"EdDSA":  jwa.EdDSA,
	}
)

// VerifyDIDJWS verifies a JWS signature against the DID specified in the JWS header object.
// It is given a did.Registry to resolve the DID Document.
func VerifyDIDJWS(ctx context.Context, registry did.Registry, message []byte) (bool, error) {

	parsedJWS, err := jws.Parse(message)
	if err != nil {
		return false, err
	}

	// get the kid, ensure its a DID, and try to resolve using the registry

	// parse the signature from the jws message. If there are none, return error.
	// @attn: Do we care if there are more than 1?
	sigs := parsedJWS.Signatures()
	if len(sigs) == 0 {
		return false, ErrMissingSignatures
	}

	alg := sigs[0].ProtectedHeaders().Algorithm()
	if _, ok := supportedJWA[alg]; !ok {
		return false, fmt.Errorf("given alg %v: %w", alg, ErrUnsupportedJWSAlgorithm)
	}

	kid := sigs[0].ProtectedHeaders().KeyID()
	if !strings.HasPrefix(kid, "did:") {
		return false, fmt.Errorf("given kid %v: %w", kid, ErrInvalidKID)
	}

	// split the DID identifer as <did>#<key>
	kidparts := strings.Split(kid, "#")
	if len(kidparts) != 2 {
		return false, fmt.Errorf("given kid %v: %w", kid, ErrDIDMissingKeyID)
	}

	// get the DID docuemnt from the registry
	// parse the <key> from the <did>#<key> in the VerificationMethod/PublicKey property
	// if the <key> object is a of the type Ed25519VerificationKey2018 use EdDSA,
	// if it is of the type EcdsaSecp256k1VerificationKey2019 use ES256K

	docResp, err := registry.Resolve(kidparts[0])
	if err != nil {
		return false, fmt.Errorf("failed to resolve did %v: %w", kidparts[0], err)
	}

	var ok bool
	var sigAlg jwa.SignatureAlgorithm
	var pubkey []byte
	for _, v := range docResp.DIDDocument.VerificationMethod {
		if v.ID == kid {
			sigAlg, ok = linkedDataSignaturesToJWA[v.Type]
			if !ok {
				return false, fmt.Errorf("found alg %v: %w", v.Type, ErrUnsupportedDIDVerificationMethodType)
			}
			pubkey = v.Value
		}
	}

	// did we find a valid verificationMethod?
	if !ok {
		return false, fmt.Errorf("couldn't find %v: %w", kid, ErrMissingDIDVerificationMethod)
	}

	// format pubkey to type
	var key interface{}
	if sigAlg == jwa.ES256K {
		secp256kPubKey, err := secp256k1.ParsePubKey(pubkey)
		if err != nil {
			return false, fmt.Errorf("failed to parse secp256k1 pubkey from DID: %w", err)
		}
		key = secp256kPubKey.ToECDSA()
	} else if sigAlg == jwa.EdDSA {
		key = ed25519.PublicKey(pubkey)
	}

	// now that we have parsed and validated all our info, we can run the signature verification
	if _, err := jws.Verify(parsedJWS.Payload(), sigAlg, key); err != nil {
		return false, fmt.Errorf("failed to verify JWS signature: %w", err)
	}

	return true, nil
}
