package blsces

import (
	_ "github.com/drand/kyber-bls12381"
	_ "github.com/drand/kyber/sign/bls"
)

// Credential is an signed set of data organized as a
// collection of Attributes and Values. IE: Attr: Value.
type Credential struct {
	CEAS       CEAS
	Signatures []Signature
	Messages   [][]byte
}

// SubCredential is an extraction from the originally
// issued Credential.
type SubCredential struct {
	CEAS CEAS
}

type Signature struct{}

// CEAS is a Content Extraction Access Structure. It is responsible
// for holding the mapping of attribute pairs and their respective
// indexes for position binding in the CES.
type CEAS struct {
}
