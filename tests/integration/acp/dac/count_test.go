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
