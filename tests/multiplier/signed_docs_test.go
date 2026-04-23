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

func TestSignedDocsName_ReturnsSignedDocs(t *testing.T) {
	sd := &signedDocs{}
	assert.Equal(t, SignedDocs, sd.Name())
	assert.Equal(t, Name("signed-docs"), sd.Name())
}

func TestSignedDocsApply_WithEmptyActions_ReturnsEmpty(t *testing.T) {
	sd := &signedDocs{}
	input := action.Actions{}

	result := sd.Apply(input)

	assert.Empty(t, result)
}

func TestSignedDocsApply_WithNilActions_ReturnsNil(t *testing.T) {
	sd := &signedDocs{}

	result := sd.Apply(nil)

	assert.Nil(t, result)
}

func TestSignedDocsApply_WithSingleAction_ReturnsIdenticalSlice(t *testing.T) {
	sd := &signedDocs{}
	add := &action.AddCollection{SDL: "type User { name: String }"}
	input := action.Actions{add}

	result := sd.Apply(input)

	assert.Len(t, result, 1)
	assert.Same(t, add, result[0], "Apply must not replace action instances")
}

func TestSignedDocsApply_WithMultipleActions_PreservesOrderAndInstances(t *testing.T) {
	sd := &signedDocs{}
	a1 := &action.AddCollection{SDL: "type User { name: String }"}
	a2 := &action.AddDoc{CollectionID: 0, Doc: `{"name": "Alice"}`}
	a3 := &action.Request{Request: `query { User { name } }`}
	input := action.Actions{a1, a2, a3}

	result := sd.Apply(input)

	assert.Len(t, result, 3)
	assert.Same(t, a1, result[0])
	assert.Same(t, a2, result[1])
	assert.Same(t, a3, result[2])
}

func TestSignedDocsApply_DoesNotMutateInputSlice(t *testing.T) {
	sd := &signedDocs{}
	a1 := &action.AddCollection{SDL: "type User { name: String }"}
	originalSDL := a1.SDL
	input := action.Actions{a1}

	_ = sd.Apply(input)

	assert.Equal(t, originalSDL, a1.SDL,
		"Apply is documented as a no-op and must not mutate action fields")
	assert.Len(t, input, 1)
	assert.Same(t, a1, input[0])
}

func TestSignedDocs_ImplementsMultiplierInterface(t *testing.T) {
	var _ Multiplier = (*signedDocs)(nil)
	var _ m.Multiplier = (*signedDocs)(nil)
}

func TestSignedDocs_DoesNotImplementActionAwareSkipper(t *testing.T) {
	// The signed-docs multiplier deliberately does NOT implement
	// [m.ActionAwareSkipper]. Skip behavior is handled:
	//   1. Automatically by the harness hook (which only upgrades
	//      EnableSigning from false to true, leaving signing-aware
	//      tests unaffected), and
	//   2. Explicitly by test authors via
	//      MultiplierExcludes: []string{multiplier.SignedDocs} for
	//      tests with structural incompatibilities such as hardcoded
	//      block CIDs.
	// This test locks in that design choice — if someone adds
	// ShouldSkip, they must update the test and the accompanying
	// decision doc.
	var sd any = &signedDocs{}
	_, ok := sd.(m.ActionAwareSkipper)
	assert.False(t, ok,
		"signedDocs must not implement ActionAwareSkipper; see decisions.md")
}

func TestSignedDocs_IsRegistered(t *testing.T) {
	// The package init() registers the multiplier with testo. We verify
	// registration indirectly by activating it via Init and asking the
	// testo package to report active multipliers.
	m.Init("__signed_docs_test_unset_env__", SignedDocs)
	t.Cleanup(func() {
		// Reset to a no-default state to avoid leaking test state into
		// other tests running in the same process.
		m.Init("__signed_docs_test_unset_env__")
	})

	active := m.Get()
	assert.Contains(t, active, string(SignedDocs))
}

func TestSignedDocsName_Stable(t *testing.T) {
	// The multiplier name is part of the public CI contract (used in
	// DEFRA_MULTIPLIERS and in MultiplierExcludes). Changing it would
	// break CI configuration and every test file that excludes it.
	assert.Equal(t, "signed-docs", string(SignedDocs))
}
