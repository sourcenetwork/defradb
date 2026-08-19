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

package action

import (
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// cfg is unused: the js build has no SourceHub container to gate on.
func setupSourceHub(s *state.State, cfg NodeSetupConfig) (*options.NodeDocumentACPOptions, error) {
	return s.DocumentACPOptions, nil
}
