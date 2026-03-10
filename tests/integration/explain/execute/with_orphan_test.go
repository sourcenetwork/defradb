// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainWithOrphanNode_WithPrimaryParent_ReportsMetrics(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},

			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},

			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},

			&action.ExplainRequest{
				// @exhaustive is required to include orphanNode in the plan
				Request: `query @explain(type: execute) @exhaustive {
					Publisher(order: {book: {rating: ASC}}) {
						name
					}
				}`,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName: "orphanNode",
						ExpectedAttributes: dataMap{
							"iterations":   uint64(2),
							"docFetches":   uint64(1),
							"fieldFetches": uint64(1),
							"indexFetches": uint64(1),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainWithOrphanNode_WithSecondaryParent_ReportsMetrics(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						publisher: Publisher
					}
					type Publisher {
						name: String
						establishedYear: Int @index
						book: Book @primary
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "OrphanBook"}`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "LinkedBook"}`,
			},

			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Publisher1",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 1),
				},
			},

			&action.ExplainRequest{
				// @exhaustive is required to include orphanNode in the plan
				Request: `query @explain(type: execute) @exhaustive {
					Book(order: {publisher: {establishedYear: ASC}}) {
						title
					}
				}`,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName: "orphanNode",
						ExpectedAttributes: dataMap{
							"iterations": uint64(3),
							// Secondary parent: orphanNode scans all parent docs (2 Books) and does
							// point lookups on the child's (Publisher) unique FK index for each.
							// The linked Book's lookup finds 1 Publisher doc.
							"docFetches":   uint64(3),
							"fieldFetches": uint64(5),
							"indexFetches": uint64(1),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
