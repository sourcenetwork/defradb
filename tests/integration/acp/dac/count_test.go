// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_QueryCountDocumentsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Request: `
					query {
						COUNT(Employee: {})
					}
				`,
				Results: map[string]any{
					"COUNT": int(2),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Request: `
					query {
						Company {
							COUNT(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							// 1 of 2 companies is public and has 1 public employee out of 2
							"COUNT": int(1),
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

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						COUNT(Employee: {})
					}
				`,
				Results: map[string]any{
					"COUNT": int(4),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Company {
							COUNT(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							"COUNT": int(2),
						},
						{
							"COUNT": int(2),
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

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						COUNT(Employee: {})
					}
				`,
				Results: map[string]any{
					"COUNT": int(2),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountRelatedObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Company {
							COUNT(employees: {})
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							// 1 of 2 companies is public and has 1 public employee out of 2
							"COUNT": int(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
