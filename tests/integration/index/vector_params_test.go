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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// An HNSW parameter above its maximum is rejected at index creation. Without the cap an oversized M
// makes the first insert do a burst of work that any client can trigger when Node-ACP is off.
func TestVectorIndex_CreateWithOversizedM_IsRejected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE, M: 100000})
				}`,
				ExpectedError: "vector index parameter is out of range",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
