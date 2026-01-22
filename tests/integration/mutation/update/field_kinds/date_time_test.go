// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithDateTimeField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"created_at": "2021-07-23T02:22:22-05:00"
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							created_at
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"created_at": testUtils.MustParseTime("2021-07-23T02:22:22-05:00"),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithDateTimeField_MultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"created_at": "2021-07-23T02:22:22-05:00"
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(input: {created_at: "2031-07-23T03:23:23Z"}) {
						name
						created_at
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":       "John",
							"created_at": testUtils.MustParseTime("2031-07-23T03:23:23Z"),
						},
						{
							"name":       "Fred",
							"created_at": testUtils.MustParseTime("2031-07-23T03:23:23Z"),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_IfDateTimeFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"created_at": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							created_at
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"created_at": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithDateTimeField_WithUTCNow(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			// Perform mutation to update using UTC_NOW
			&action.Request{
				Request: `mutation {
					update_Users(
						filter: { name: { _eq: "John" } },
						input: { created_at: UTC_NOW }
					) {
						name
						created_at
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":       "John",
							"created_at": testUtils.CurrentTimestamp(),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithDateTimeField_WithUTCNow_ShouldBeEqual(t *testing.T) {
	timestampMatcher := testUtils.NewSameValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Chris",
					"created_at": "2012-07-23T01:11:11-05:00"
				}`,
			},
			// Perform mutations to update using UTC_NOW
			&action.Request{
				Request: `mutation {
					john: update_Users(
						filter: { name: { _eq: "John" } },
						input: { created_at: UTC_NOW }
					) {
						created_at
					}
					chris: update_Users(
						filter: { name: { _eq: "Chris" } },
						input: { created_at: UTC_NOW }
					) {
						created_at
					}
				}`,
				Results: map[string]any{
					"john":  []map[string]any{{"created_at": timestampMatcher}},
					"chris": []map[string]any{{"created_at": timestampMatcher}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
