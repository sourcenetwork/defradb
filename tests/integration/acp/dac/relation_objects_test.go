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

func TestACP_QueryManyToOneRelationObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Employee": []map[string]any{
						{
							"name":    "PubEmp in PubCompany",
							"company": map[string]any{"name": "Public Company"},
						},
						{
							"name":    "PubEmp in PrivateCompany",
							"company": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							"name": "Public Company",
							"employees": []map[string]any{
								{"name": "PubEmp in PubCompany"},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryManyToOneRelationObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Employee": []map[string]any{
						{
							"name":    "PrivateEmp in PrivateCompany",
							"company": map[string]any{"name": "Private Company"},
						},
						{
							"name":    "PubEmp in PubCompany",
							"company": map[string]any{"name": "Public Company"},
						},
						{
							"name":    "PubEmp in PrivateCompany",
							"company": map[string]any{"name": "Private Company"},
						},
						{
							"name":    "PrivateEmp in PubCompany",
							"company": map[string]any{"name": "Public Company"},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							"name": "Public Company",
							"employees": []map[string]any{
								{"name": "PubEmp in PubCompany"},
								{"name": "PrivateEmp in PubCompany"},
							},
						},
						{
							"name": "Private Company",
							"employees": []map[string]any{
								{"name": "PrivateEmp in PrivateCompany"},
								{"name": "PubEmp in PrivateCompany"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryManyToOneRelationObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Employee": []map[string]any{
						{
							"name":    "PubEmp in PubCompany",
							"company": map[string]any{"name": "Public Company"},
						},
						{
							"name":    "PubEmp in PrivateCompany",
							"company": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"Company": []map[string]any{
						{
							"name": "Public Company",
							"employees": []map[string]any{
								{"name": "PubEmp in PubCompany"},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
