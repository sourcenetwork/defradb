package identity

import (
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
)

func Test_DIDFromPublicKey_ProducesDIDForPublicKey(t *testing.T) {
	pubKey := &secp256k1.PublicKey{}

	did := DIDFromPublicKey(pubKey)

	want := "did:key:z7r8ooUiNXK8TT8Xjg1EWStR2ZdfxbzVfvGWbA2FjmzcnmDxz71QkP1Er8PP3zyLZpBLVgaXbZPGJPS4ppXJDPRcqrx4F"
	require.Equal(t, want, did)
}
