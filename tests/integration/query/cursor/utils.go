// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cursor

import (
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// userCollectionGQLSchema defines a User type with indexed age field for cursor tests.
var userCollectionGQLSchema = `
	type User {
		name: String
		age: Int @index
	}
`

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			SupportedMutationTypes: test.SupportedMutationTypes,
			SupportedClientTypes:   test.SupportedClientTypes,
			Actions: append(
				[]any{
					&action.AddCollection{
						SDL: userCollectionGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}

// makeExplainQuery wraps a query with @explain(type: execute).
func makeExplainQuery(req string) string {
	ind := strings.Index(req, "query")
	if ind < 0 {
		panic("Invalid query: " + req)
	}
	return "query @explain(type: execute) " + req[ind+5:]
}

// dynamicRequest generates its request string at execution time, enabling
// cursor capture from one query to be used in a subsequent query.
type dynamicRequest struct {
	s         *state.State
	requestFn func() string
	asserter  testUtils.ResultAsserter
}

var _ action.Stateful = (*dynamicRequest)(nil)

func (d *dynamicRequest) SetState(s *state.State) {
	d.s = s
}

func (d *dynamicRequest) Execute() {
	request := d.requestFn()

	if len(d.s.Nodes) == 0 {
		d.s.T.Fatal("no nodes available for request execution")
		return
	}

	node := d.s.Nodes[0]
	result := node.ExecRequest(d.s.Ctx, request)

	if len(result.GQL.Errors) > 0 {
		d.s.T.Fatalf("request failed with errors: %v", result.GQL.Errors)
		return
	}

	if d.asserter != nil {
		data, ok := result.GQL.Data.(map[string]any)
		if !ok {
			d.s.T.Fatalf("expected map[string]any, got %T", result.GQL.Data)
			return
		}
		d.asserter.Assert(d.s.T, data)
	}
}
