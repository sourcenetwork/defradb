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
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

func TestACP_QueryRelationObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees and companies without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
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
				Results: []map[string]any{
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

			testUtils.Request{
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
				Results: []map[string]any{
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PubEmp in PubCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryRelationObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees and companies with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
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
				Results: []map[string]any{
					{
						"name":    "PubEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PrivateEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PrivateEmp in PrivateCompany",
						"company": map[string]any{"name": "Private Company"},
					},
					{
						"name":    "PubEmp in PrivateCompany",
						"company": map[string]any{"name": "Private Company"},
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
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
				Results: []map[string]any{
					{
						"name": "Private Company",
						"employees": []map[string]any{
							{"name": "PrivateEmp in PrivateCompany"},
							{"name": "PubEmp in PrivateCompany"},
						},
					},
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PubEmp in PubCompany"},
							{"name": "PrivateEmp in PubCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryRelationObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees and companies with wrong identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
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
				Results: []map[string]any{
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

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
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
				Results: []map[string]any{
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PubEmp in PubCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
