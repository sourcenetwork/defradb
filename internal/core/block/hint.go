// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"crypto/hmac"
	"crypto/sha256"
)

// EncryptionHintLength is the byte length of the truncated HMAC hint.
// 16 bytes (128 bits) provides sufficient collision resistance for
// filtering purposes while keeping block size overhead minimal.
const EncryptionHintLength = 16

// GenerateEncryptionHint generates a truncated HMAC hint for quick key identification.
// The hint is computed as HMAC-SHA256(key, docID || fieldName)[:16].
//
// This hint can be used to quickly check if a local key might decrypt an encrypted
// block without fetching the full Encryption block from storage.
func GenerateEncryptionHint(key []byte, docID []byte, fieldName []byte) []byte {
	if len(key) == 0 {
		return nil
	}

	var data []byte
	data = append(data, docID...)
	data = append(data, fieldName...)

	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	fullHMAC := mac.Sum(nil)

	return fullHMAC[:EncryptionHintLength]
}

// MatchesEncryptionHint checks if the given key would produce the same hint.
// This is a quick local check to determine if we likely have the decryption key.
//
// Returns true if the hint matches, indicating a high probability that the
// provided key can decrypt the block. Returns false if hint is nil/invalid
// or if the computed hint doesn't match.
func MatchesEncryptionHint(hint []byte, key []byte, docID []byte, fieldName []byte) bool {
	if len(hint) != EncryptionHintLength {
		return false
	}
	if len(key) == 0 {
		return false
	}
	expectedHint := GenerateEncryptionHint(key, docID, fieldName)
	return hmac.Equal(hint, expectedHint)
}
