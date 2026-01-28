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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const employeeCompanyPolicy = `
description: A Policy
name: test
resources:
- name: companies
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
- name: employees
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
`

func getSetupEmployeeCompanyActions() []any {
	return []any{
		testUtils.AddDACPolicy{
			Identity: testUtils.ClientIdentity(1),
			Policy:   employeeCompanyPolicy,
		},

		&action.AddSchema{
			Schema: `
					type Employee @policy(
						id: "{{.Policy0}}",
						resource: "employees"
					) {
						name: String
						salary: Int
						company: Company
					}

					type Company @policy(
						id: "{{.Policy0}}",
						resource: "companies"
					) {
						name: String
						capital: Int
						employees: [Employee]
					}
				`,
		},

		&action.CreateDoc{
			CollectionID: 1,
			Doc: `
					{
						"name": "Public Company",
						"capital": 100000
					}
				`,
		},
		&action.CreateDoc{
			CollectionID: 1,
			Identity:     testUtils.ClientIdentity(1),
			Doc: `
					{
						"name": "Private Company",
						"capital": 200000
					}
				`,
		},
		&action.CreateDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PubCompany",
				"salary":  10000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		&action.CreateDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PrivateCompany",
				"salary":  20000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
		&action.CreateDoc{
			CollectionID: 0,
			Identity:     testUtils.ClientIdentity(1),
			DocMap: map[string]any{
				"name":    "PrivateEmp in PubCompany",
				"salary":  30000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		&action.CreateDoc{
			CollectionID: 0,
			Identity:     testUtils.ClientIdentity(1),
			DocMap: map[string]any{
				"name":    "PrivateEmp in PrivateCompany",
				"salary":  40000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
	}
}
