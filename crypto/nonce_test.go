// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNonceFunc_UsesDeterministicInTest(t *testing.T) {
	isTesting := testing.Testing()
	t.Logf("testing.Testing() returned: %v", isTesting)

	// Generate two nonces
	nonce1, err := generateNonceFunc()
	assert.NoError(t, err)

	nonce2, err := generateNonceFunc()
	assert.NoError(t, err)

	// In test mode, they should be identical (deterministic)
	// In production mode, they should be different (random)
	if isTesting {
		assert.Equal(t, nonce1, nonce2, "Nonces should be deterministic in test mode")
		t.Logf("Nonces are deterministic (test mode confirmed)")
	} else {
		assert.NotEqual(t, nonce1, nonce2, "Nonces should be random in production")
		t.Logf("Nonces are random (production mode confirmed)")
	}
}
