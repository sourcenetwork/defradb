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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

const employeeCompanyPolicy = `
name: test
description: A Valid DefraDB Policy Interface (DPI)

actor:
  name: actor

resources:
  employees:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor

  companies:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
`

func getSetupEmployeeCompanyActions() []any {
	return []any{
		testUtils.AddPolicy{
			Identity:         acpUtils.Actor1Identity,
			Policy:           employeeCompanyPolicy,
			ExpectedPolicyID: "9d6c19007a894746c3f45f7fe45513a88a20ad77637948228869546197bb1b05",
		},

		testUtils.SchemaUpdate{
			Schema: `
					type Employee @policy(
						id: "9d6c19007a894746c3f45f7fe45513a88a20ad77637948228869546197bb1b05",
						resource: "employees"
					) {
						name: String
						salary: Int
						company: Company
					}

					type Company @policy(
						id: "9d6c19007a894746c3f45f7fe45513a88a20ad77637948228869546197bb1b05",
						resource: "companies"
					) {
						name: String
						capital: Int
						employees: [Employee]
					}
				`,
		},

		testUtils.CreateDoc{
			CollectionID: 1,
			Doc: `
					{
						"name": "Public Company",
						"capital": 100000
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 1,
			Identity:     acpUtils.Actor1Identity,
			Doc: `
					{
						"name": "Private Company",
						"capital": 200000
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PubCompany",
				"salary":  10000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PrivateCompany",
				"salary":  20000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Identity:     acpUtils.Actor1Identity,
			DocMap: map[string]any{
				"name":    "PrivateEmp in PubCompany",
				"salary":  30000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Identity:     acpUtils.Actor1Identity,
			DocMap: map[string]any{
				"name":    "PrivateEmp in PrivateCompany",
				"salary":  40000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
	}
}
