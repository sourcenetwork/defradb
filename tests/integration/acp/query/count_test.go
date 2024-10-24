// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_QueryCountDocumentsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query documents' count without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: map[string]any{
					"_count": int(2),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count of related objects without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							// 1 of 2 companies is public and has 1 public employee out of 2
							"_count": int(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountDocumentsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query documents' count with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: testUtils.UserIdentity(1),
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: map[string]any{
					"_count": int(4),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count of related objects with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: testUtils.UserIdentity(1),
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							"_count": int(2),
						},
						{
							"_count": int(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountDocumentsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query documents' count without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: testUtils.UserIdentity(2),
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: map[string]any{
					"_count": int(2),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count of related objects without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: testUtils.UserIdentity(2),
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							// 1 of 2 companies is public and has 1 public employee out of 2
							"_count": int(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
