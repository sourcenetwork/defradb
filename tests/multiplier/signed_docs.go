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
	multiplier.Register(&signedDocs{})
}

// SignedDocs multiplier automatically enables document signing
// ([TestCase.EnableSigning]) on tests that do not already configure it.
//
// Signing adds a [Signature] link to composite blocks and priority-1 field
// blocks in the Merkle DAG, which changes block bytes (and therefore CIDs).
// This multiplier verifies that every feature in the test suite behaves
// correctly under signing.
//
// Unlike other multipliers, [signedDocs.Apply] is a no-op. The harness applies
// this multiplier at TestCase level, enabling signing and keeping generated
// create actions unsigned unless a test opted into signing itself.
//
// Tests that cannot cope with signing (e.g. hardcoded block CIDs in result
// assertions) must opt out via
// MultiplierExcludes: []string{multiplier.SignedDocs}.
const SignedDocs Name = "signed-docs"

type signedDocs struct{}

var _ Multiplier = (*signedDocs)(nil)

func (m *signedDocs) Name() Name {
	return SignedDocs
}

// Apply is a no-op. The signed-docs multiplier's effect is applied by the
// integration harness because it needs access to TestCase-level settings.
func (m *signedDocs) Apply(source action.Actions) action.Actions {
	return source
}
