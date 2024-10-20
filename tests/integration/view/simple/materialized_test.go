// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleMaterialized_DoesNotAutoUpdateOnViewCreate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			// As the MaterializedViewType will auto refresh views immediately prior
			// to executing requests, this test of materialized views actually only
			// supports running with the CachelessViewType flag.
			testUtils.CachelessViewType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					// Even though UserView was created after the document was created, the results are
					// empty because the view will not populate until RefreshView is called.
					"UserView": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleMaterialized_DoesNotAutoUpdate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			// As the MaterializedViewType will auto refresh views immediately prior
			// to executing requests, this test of materialized views actually only
			// supports running with the CachelessViewType flag.
			testUtils.CachelessViewType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.RefreshViews{},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
