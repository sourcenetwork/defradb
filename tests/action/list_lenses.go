// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/state"
)

// ListLenses is an action that lists all stored lenses and optionally verifies them.
type ListLenses struct {
	stateful

	// NodeID may hold the ID (index) of a node to list lenses from.
	//
	// If a value is not provided the lenses will be listed from all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	Identity immutable.Option[state.Identity]

	// ExpectedLenses is a map of lens CID to expected lens configuration.
	// Keys can use template syntax (e.g., "{{.LensID0}}") which will be resolved
	// to actual CIDs at execution time.
	// If set, the action will verify the lens content matches.
	ExpectedLenses map[string]model.Lens
}

var _ Action = (*ListLenses)(nil)
var _ Stateful = (*ListLenses)(nil)

func (a *ListLenses) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		a.s.Ctx = getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)

		lenses, err := node.ListLenses(a.s.Ctx)

		resetStateContext(a.s)

		if err != nil {
			a.s.T.Fatalf("failed to list lenses: %v", err)
		}

		if a.ExpectedLenses == nil {
			return
		}

		templateKeys := make([]string, 0, len(a.ExpectedLenses))
		for key := range a.ExpectedLenses {
			templateKeys = append(templateKeys, key)
		}
		resolvedKeys := replaceMap(a.s, nodeID, templateKeys)

		assert.Equal(a.s.T, len(a.ExpectedLenses), len(lenses),
			"expected %d lenses, got %d", len(a.ExpectedLenses), len(lenses))

		// We compare module count, arguments, and inverse flag, but not the Path field
		// because when stored, the Path changes from a file path to embedded WASM data.
		for templateKey, expectedLens := range a.ExpectedLenses {
			lensID := resolvedKeys[templateKey]

			actualLens, exists := lenses[lensID]
			require.True(a.s.T, exists, "expected lens %s (resolved from %s) not found", lensID, templateKey)

			require.Equal(a.s.T, len(expectedLens.Lenses), len(actualLens.Lenses),
				"lens module count mismatch for lens %s", lensID)

			for i, expectedModule := range expectedLens.Lenses {
				actualModule := actualLens.Lenses[i]
				assert.Equal(a.s.T, expectedModule.Inverse, actualModule.Inverse,
					"lens module[%d] inverse mismatch for lens %s", i, lensID)
				assert.Equal(a.s.T, expectedModule.Arguments, actualModule.Arguments,
					"lens module[%d] arguments mismatch for lens %s", i, lensID)
			}
		}
	}
}
