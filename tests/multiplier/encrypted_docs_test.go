// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package multiplier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	m "github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/tests/action"
)

func TestEncryptedDocsName_ReturnsEncryptedDocs(t *testing.T) {
	ed := &encryptedDocs{}
	assert.Equal(t, EncryptedDocs, ed.Name())
	assert.Equal(t, Name("encrypted-docs"), ed.Name())
}

func TestEncryptedDocsName_Stable(t *testing.T) {
	// The multiplier name is part of the public CI contract (used in
	// DEFRA_MULTIPLIERS and in MultiplierExcludes). Changing it would
	// break CI configuration and every test file that excludes it.
	assert.Equal(t, "encrypted-docs", string(EncryptedDocs))
}

func TestEncryptedDocsApply_WithEmptyActions_ReturnsEmpty(t *testing.T) {
	ed := &encryptedDocs{}
	input := action.Actions{}

	result := ed.Apply(input)

	assert.Empty(t, result)
}

func TestEncryptedDocsApply_WithNilActions_ReturnsNil(t *testing.T) {
	ed := &encryptedDocs{}

	result := ed.Apply(nil)

	assert.Nil(t, result)
}

func TestEncryptedDocsApply_FlipsPlainAddDoc(t *testing.T) {
	ed := &encryptedDocs{}
	add := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`}
	input := action.Actions{add}

	result := ed.Apply(input)

	assert.Len(t, result, 1)
	assert.Same(t, add, result[0])
	assert.True(t, add.IsDocEncrypted, "Apply must flip IsDocEncrypted on un-configured AddDoc")
	assert.Empty(t, add.EncryptedFields, "Apply must not populate EncryptedFields")
}

func TestEncryptedDocsApply_PreservesEncryptedAddDoc(t *testing.T) {
	// In practice ShouldSkip will skip the whole test before Apply runs,
	// but Apply must still be safe if called directly.
	ed := &encryptedDocs{}
	add := &action.AddDoc{
		CollectionID:   0,
		Doc:            `{"name": "Alice"}`,
		IsDocEncrypted: true,
	}
	input := action.Actions{add}

	_ = ed.Apply(input)

	assert.True(t, add.IsDocEncrypted)
	assert.Empty(t, add.EncryptedFields)
}

func TestEncryptedDocsApply_PreservesPartialFieldEncryption(t *testing.T) {
	ed := &encryptedDocs{}
	add := &action.AddDoc{
		CollectionID:    0,
		Doc:             `{"name": "Alice"}`,
		EncryptedFields: []string{"name"},
	}
	input := action.Actions{add}

	_ = ed.Apply(input)

	assert.False(t, add.IsDocEncrypted,
		"Apply must NOT escalate field-level encryption to doc-level encryption")
	assert.Equal(t, []string{"name"}, add.EncryptedFields)
}

func TestEncryptedDocsApply_LeavesNonAddDocUntouched(t *testing.T) {
	ed := &encryptedDocs{}
	a1 := &action.AddCollection{SDL: "type User { name: String }"}
	a2 := &action.Request{Request: `query { User { name } }`}
	originalSDL := a1.SDL
	originalRequest := a2.Request
	input := action.Actions{a1, a2}

	result := ed.Apply(input)

	assert.Len(t, result, 2)
	assert.Same(t, a1, result[0])
	assert.Same(t, a2, result[1])
	assert.Equal(t, originalSDL, a1.SDL)
	assert.Equal(t, originalRequest, a2.Request)
}

func TestEncryptedDocsApply_MultipleAddDocs_MixedConfig(t *testing.T) {
	ed := &encryptedDocs{}
	plain := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`}
	alreadyEncrypted := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Bob"}`, IsDocEncrypted: true}
	fieldEncrypted := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Carol"}`, EncryptedFields: []string{"name"}}
	input := action.Actions{plain, alreadyEncrypted, fieldEncrypted}

	result := ed.Apply(input)

	assert.Len(t, result, 3)
	assert.True(t, plain.IsDocEncrypted, "un-configured AddDoc must be flipped")
	assert.True(t, alreadyEncrypted.IsDocEncrypted, "already-encrypted AddDoc must remain encrypted")
	assert.False(t, fieldEncrypted.IsDocEncrypted,
		"field-encrypted AddDoc must NOT be escalated to doc-level encryption")
	assert.Equal(t, []string{"name"}, fieldEncrypted.EncryptedFields)
}

func TestEncryptedDocsApply_PreservesOrderAndInstances(t *testing.T) {
	ed := &encryptedDocs{}
	a1 := &action.AddCollection{SDL: "type User { name: String }"}
	a2 := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`}
	a3 := &action.Request{Request: `query { User { name } }`}
	input := action.Actions{a1, a2, a3}

	result := ed.Apply(input)

	assert.Len(t, result, 3)
	assert.Same(t, a1, result[0])
	assert.Same(t, a2, result[1])
	assert.Same(t, a3, result[2])
}

func TestEncryptedDocsShouldSkip_WithEncryptedAddDoc_ReturnsTrue(t *testing.T) {
	ed := &encryptedDocs{}
	input := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`, IsDocEncrypted: true},
	}

	assert.True(t, ed.ShouldSkip(input))
}

func TestEncryptedDocsShouldSkip_WithFieldEncryptedAddDoc_ReturnsTrue(t *testing.T) {
	ed := &encryptedDocs{}
	input := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`, EncryptedFields: []string{"name"}},
	}

	assert.True(t, ed.ShouldSkip(input))
}

func TestEncryptedDocsShouldSkip_WithMixedAddDocs_ReturnsTrue(t *testing.T) {
	// Even one encryption-configured AddDoc anywhere in the test triggers a
	// whole-test skip. Tests that mix configured and unconfigured AddDocs
	// are deliberately encryption-aware and their intent must be preserved.
	ed := &encryptedDocs{}
	input := action.Actions{
		&action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`},
		&action.AddDoc{CollectionID: 0, Doc: `{"name": "Bob"}`, IsDocEncrypted: true},
	}

	assert.True(t, ed.ShouldSkip(input))
}

func TestEncryptedDocsShouldSkip_WithOrdinaryActions_ReturnsFalse(t *testing.T) {
	ed := &encryptedDocs{}
	input := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`},
		&action.Request{Request: `query { User { name } }`},
	}

	assert.False(t, ed.ShouldSkip(input))
}

func TestEncryptedDocsShouldSkip_WithNoAddDoc_ReturnsFalse(t *testing.T) {
	ed := &encryptedDocs{}
	input := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.Request{Request: `query { User { name } }`},
	}

	assert.False(t, ed.ShouldSkip(input))
}

func TestEncryptedDocsShouldSkip_WithEmptyActions_ReturnsFalse(t *testing.T) {
	ed := &encryptedDocs{}

	assert.False(t, ed.ShouldSkip(action.Actions{}))
}

func TestEncryptedDocs_ImplementsMultiplierInterface(t *testing.T) {
	var _ Multiplier = (*encryptedDocs)(nil)
	var _ m.Multiplier = (*encryptedDocs)(nil)
}

func TestEncryptedDocs_ImplementsActionAwareSkipper(t *testing.T) {
	var ed any = &encryptedDocs{}
	_, ok := ed.(m.ActionAwareSkipper)
	assert.True(t, ok,
		"encryptedDocs must implement ActionAwareSkipper to skip encryption-aware tests")
}

func TestEncryptedDocs_IsRegistered(t *testing.T) {
	m.Init("__encrypted_docs_test_unset_env__", EncryptedDocs)
	t.Cleanup(func() {
		m.Init("__encrypted_docs_test_unset_env__")
	})

	active := m.Get()
	assert.Contains(t, active, string(EncryptedDocs))
}
