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
	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/tests/action"
)

func init() {
	multiplier.Register(&encryptedDocs{})
}

// EncryptedDocs is the name of the multiplier that automatically enables
// document encryption ([action.AddDoc.IsDocEncrypted]) on AddDoc actions
// of tests that do not already configure encryption.
//
// Encryption changes block payload bytes (and therefore CIDs), and runs a
// key-management path on every read/merge. This multiplier verifies that
// every feature in the test suite behaves correctly when documents are
// encrypted.
//
// Tests that cannot cope with encryption (e.g. hardcoded block CIDs in
// result assertions, byte-level block-content assertions) must opt out via
// MultiplierExcludes: []string{multiplier.EncryptedDocs}.
const EncryptedDocs Name = "encrypted-docs"

type encryptedDocs struct{}

var _ Multiplier = (*encryptedDocs)(nil)
var _ multiplier.ActionAwareSkipper = (*encryptedDocs)(nil)

func (m *encryptedDocs) Name() Name {
	return EncryptedDocs
}

// Apply rewrites every [*action.AddDoc] that has not already configured
// encryption (i.e. IsDocEncrypted is false AND EncryptedFields is empty)
// to set IsDocEncrypted = true. Other actions are returned unchanged.
//
// The harness's toAction conversion produces pointers for AddDoc, so
// in-place mutation of the pointed-to struct is observable by the harness
// and survives into action execution.
func (m *encryptedDocs) Apply(source action.Actions) action.Actions {
	for _, a := range source {
		addDoc, ok := a.(*action.AddDoc)
		if !ok {
			continue
		}
		if addDoc.IsDocEncrypted || len(addDoc.EncryptedFields) > 0 {
			continue
		}
		addDoc.IsDocEncrypted = true
	}
	return source
}

// ShouldSkip implements [multiplier.ActionAwareSkipper].
//
// Skips tests where any [*action.AddDoc] already configures encryption
// (either IsDocEncrypted or EncryptedFields). Such tests are
// encryption-aware by construction and their intent must be preserved
// verbatim — flipping the "other" AddDoc actions in the same test would
// silently invalidate the test's assertions.
//
// Other structural incompatibilities (hardcoded block CIDs, byte-level
// block-content assertions, etc.) are handled via explicit
// MultiplierExcludes: []string{multiplier.EncryptedDocs} on the test case,
// NOT via auto-detection here. Explicit opt-out is more discoverable and
// forces the author to document the incompatibility.
func (m *encryptedDocs) ShouldSkip(actions action.Actions) bool {
	for _, a := range actions {
		addDoc, ok := a.(*action.AddDoc)
		if !ok {
			continue
		}
		if addDoc.IsDocEncrypted || len(addDoc.EncryptedFields) > 0 {
			return true
		}
	}
	return false
}
