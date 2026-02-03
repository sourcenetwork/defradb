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
	"bytes"
	"testing"
)

func TestGenerateEncryptionHint_WithValidInputs_ReturnsHint(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint := GenerateEncryptionHint(key, docID, fieldName)

	if hint == nil {
		t.Fatal("expected hint to be non-nil")
	}
	if len(hint) != EncryptionHintLength {
		t.Fatalf("expected hint length %d, got %d", EncryptionHintLength, len(hint))
	}
}

func TestGenerateEncryptionHint_WithEmptyKey_ReturnsNil(t *testing.T) {
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint := GenerateEncryptionHint(nil, docID, fieldName)

	if hint != nil {
		t.Fatal("expected nil hint for empty key")
	}
}

func TestGenerateEncryptionHint_SameInputs_ReturnsSameHint(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint1 := GenerateEncryptionHint(key, docID, fieldName)
	hint2 := GenerateEncryptionHint(key, docID, fieldName)

	if !bytes.Equal(hint1, hint2) {
		t.Fatal("expected same inputs to produce same hint")
	}
}

func TestGenerateEncryptionHint_DifferentKeys_ReturnsDifferentHints(t *testing.T) {
	key1 := []byte("test-encryption-key-32-bytes!!")
	key2 := []byte("different-encryption-key-32by!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint1 := GenerateEncryptionHint(key1, docID, fieldName)
	hint2 := GenerateEncryptionHint(key2, docID, fieldName)

	if bytes.Equal(hint1, hint2) {
		t.Fatal("expected different keys to produce different hints")
	}
}

func TestGenerateEncryptionHint_DifferentDocIDs_ReturnsDifferentHints(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID1 := []byte("bafyreib1...")
	docID2 := []byte("bafyreib2...")
	fieldName := []byte("name")

	hint1 := GenerateEncryptionHint(key, docID1, fieldName)
	hint2 := GenerateEncryptionHint(key, docID2, fieldName)

	if bytes.Equal(hint1, hint2) {
		t.Fatal("expected different docIDs to produce different hints")
	}
}

func TestGenerateEncryptionHint_DifferentFieldNames_ReturnsDifferentHints(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName1 := []byte("name")
	fieldName2 := []byte("email")

	hint1 := GenerateEncryptionHint(key, docID, fieldName1)
	hint2 := GenerateEncryptionHint(key, docID, fieldName2)

	if bytes.Equal(hint1, hint2) {
		t.Fatal("expected different fieldNames to produce different hints")
	}
}

func TestGenerateEncryptionHint_WithEmptyFieldName_ProducesValidHint(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")

	hint := GenerateEncryptionHint(key, docID, nil)

	if hint == nil {
		t.Fatal("expected hint to be non-nil")
	}
	if len(hint) != EncryptionHintLength {
		t.Fatalf("expected hint length %d, got %d", EncryptionHintLength, len(hint))
	}
}

func TestMatchesEncryptionHint_WithMatchingInputs_ReturnsTrue(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint := GenerateEncryptionHint(key, docID, fieldName)
	matches := MatchesEncryptionHint(hint, key, docID, fieldName)

	if !matches {
		t.Fatal("expected matching inputs to return true")
	}
}

func TestMatchesEncryptionHint_WithDifferentKey_ReturnsFalse(t *testing.T) {
	key1 := []byte("test-encryption-key-32-bytes!!")
	key2 := []byte("different-encryption-key-32by!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint := GenerateEncryptionHint(key1, docID, fieldName)
	matches := MatchesEncryptionHint(hint, key2, docID, fieldName)

	if matches {
		t.Fatal("expected different key to return false")
	}
}

func TestMatchesEncryptionHint_WithInvalidHintLength_ReturnsFalse(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	// Test with short hint
	matches := MatchesEncryptionHint([]byte("short"), key, docID, fieldName)
	if matches {
		t.Fatal("expected short hint to return false")
	}

	// Test with long hint
	longHint := make([]byte, EncryptionHintLength+10)
	matches = MatchesEncryptionHint(longHint, key, docID, fieldName)
	if matches {
		t.Fatal("expected long hint to return false")
	}

	// Test with nil hint
	matches = MatchesEncryptionHint(nil, key, docID, fieldName)
	if matches {
		t.Fatal("expected nil hint to return false")
	}
}

func TestMatchesEncryptionHint_WithEmptyKey_ReturnsFalse(t *testing.T) {
	key := []byte("test-encryption-key-32-bytes!!")
	docID := []byte("bafyreib...")
	fieldName := []byte("name")

	hint := GenerateEncryptionHint(key, docID, fieldName)
	matches := MatchesEncryptionHint(hint, nil, docID, fieldName)

	if matches {
		t.Fatal("expected empty key to return false")
	}
}

func TestEncryption_GenerateHint_ProducesValidHint(t *testing.T) {
	fieldName := "name"
	enc := &Encryption{
		DocID:     []byte("bafyreib..."),
		FieldName: &fieldName,
		Key:       []byte("test-encryption-key-32-bytes!!"),
	}

	hint := enc.GenerateHint()

	if hint == nil {
		t.Fatal("expected hint to be non-nil")
	}
	if len(hint) != EncryptionHintLength {
		t.Fatalf("expected hint length %d, got %d", EncryptionHintLength, len(hint))
	}
}

func TestEncryption_GenerateHint_WithNoFieldName_ProducesValidHint(t *testing.T) {
	enc := &Encryption{
		DocID: []byte("bafyreib..."),
		Key:   []byte("test-encryption-key-32-bytes!!"),
	}

	hint := enc.GenerateHint()

	if hint == nil {
		t.Fatal("expected hint to be non-nil")
	}
	if len(hint) != EncryptionHintLength {
		t.Fatalf("expected hint length %d, got %d", EncryptionHintLength, len(hint))
	}
}

func TestEncryption_MatchesHint_WithMatchingHint_ReturnsTrue(t *testing.T) {
	fieldName := "name"
	enc := &Encryption{
		DocID:     []byte("bafyreib..."),
		FieldName: &fieldName,
		Key:       []byte("test-encryption-key-32-bytes!!"),
	}

	hint := enc.GenerateHint()
	matches := enc.MatchesHint(hint)

	if !matches {
		t.Fatal("expected encryption block to match its own hint")
	}
}

func TestEncryption_MatchesHint_WithDifferentHint_ReturnsFalse(t *testing.T) {
	fieldName := "name"
	enc := &Encryption{
		DocID:     []byte("bafyreib..."),
		FieldName: &fieldName,
		Key:       []byte("test-encryption-key-32-bytes!!"),
	}

	// Create a different encryption block
	differentKey := []byte("different-encryption-key-32by!!")
	differentEnc := &Encryption{
		DocID:     []byte("bafyreib..."),
		FieldName: &fieldName,
		Key:       differentKey,
	}
	differentHint := differentEnc.GenerateHint()

	matches := enc.MatchesHint(differentHint)

	if matches {
		t.Fatal("expected encryption block to not match different hint")
	}
}

func TestBlock_HasEncryptionHint_WithValidHint_ReturnsTrue(t *testing.T) {
	block := &Block{
		EncryptionHint: make([]byte, EncryptionHintLength),
	}

	if !block.HasEncryptionHint() {
		t.Fatal("expected block with valid hint to return true")
	}
}

func TestBlock_HasEncryptionHint_WithNoHint_ReturnsFalse(t *testing.T) {
	block := &Block{}

	if block.HasEncryptionHint() {
		t.Fatal("expected block with no hint to return false")
	}
}

func TestBlock_HasEncryptionHint_WithInvalidLengthHint_ReturnsFalse(t *testing.T) {
	block := &Block{
		EncryptionHint: []byte("too-short"),
	}

	if block.HasEncryptionHint() {
		t.Fatal("expected block with invalid length hint to return false")
	}
}
