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

package branchable_collection

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// Clients built upon the C bindings serializes a context's deadline into a
// string that crosses the cgo boundary, rather than passing the context
// itself. An absent deadline (context.Background(), the common case) and a
// present one are two distinct code paths on that side. These tests are limited to
// the C client, because they are meant to test the working serialization of the
// deadlines.

// This tests serializing a blank deadline. In the C wrapper, the context will be
// examined for a deadline, and when none is present, a blank string will be passed in.
// That is the core of what this test is making sure works.
func TestBranchableCollectionSync_NoDeadlineContext_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.CClientType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID:     1,
				NoDeadline: true,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This one tests that a finite deadline is properly serialized. In this case, the wrapper
// will see a deadline on the context, and will pass it in as a string.
func TestBranchableCollectionSync_FiniteDeadlineContext_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.CClientType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID:     1,
				NoDeadline: false,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
