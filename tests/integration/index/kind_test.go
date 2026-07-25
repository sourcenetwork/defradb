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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexKind_IfKindHasNoImplementation_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM)
					}
				`,
				ExpectedError: db.NewErrUnknownIndexKind(client.IndexKindTrigram).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexKind_IfKindWithOptionsHasNoImplementation_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25, options: {k1: 1.2, b: 0.75})
					}
				`,
				ExpectedError: db.NewErrUnknownIndexKind(client.IndexKindBM25).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
