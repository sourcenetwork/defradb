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

// SDL definition order determines CollectionID:
//
//	CollectionID 0 = Employee (first type in SDL)
//	CollectionID 1 = Company  (second type in SDL)
const employeeCompanyRelationSDL = `
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
`

// policyAndSchemaSetup returns the DAC policy + AddCollection actions shared by all tests
// in this file.
func policyAndSchemaSetup() []any {
	return []any{
		testUtils.AddDACPolicy{
			Identity: testUtils.ClientIdentity(1),
			Policy:   employeeCompanyPolicy,
		},
		&action.AddCollection{
			SDL: employeeCompanyRelationSDL,
		},
	}
}

// TestACP_MutationAdd_RelationTarget_PrivateDoc_NoIdentity_Error asserts that a caller with
// no identity cannot create a document whose relation field points to a private document.
func TestACP_MutationAdd_RelationTarget_PrivateDoc_NoIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			// Identity 1 creates a private Company (CollectionID 1, index 0).
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// No-identity caller tries to create an Employee linked to the private Company.
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":    "Employee",
					"salary":  50000,
					"company": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "relation target document not found",
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationAdd_RelationTarget_PrivateDoc_WrongIdentity_Error asserts that a caller
// whose identity has no read permission on the target document cannot link to it.
func TestACP_MutationAdd_RelationTarget_PrivateDoc_WrongIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// Identity 2 has no read access to the private Company.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				DocMap: map[string]any{
					"name":    "Employee",
					"salary":  50000,
					"company": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "relation target document not found",
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationAdd_RelationTarget_PrivateDoc_OwnerIdentity_NoError asserts that the
// document owner can use their own private document as a relation target.
func TestACP_MutationAdd_RelationTarget_PrivateDoc_OwnerIdentity_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// Identity 1 owns the private Company — linking is allowed.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name":    "Employee",
					"salary":  50000,
					"company": testUtils.NewDocIndex(1, 0),
				},
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationAdd_RelationTarget_PrivateDoc_GrantedRead_NoError asserts that a caller
// who has been explicitly granted reader access can link to the private document.
func TestACP_MutationAdd_RelationTarget_PrivateDoc_GrantedRead_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// Identity 1 grants Identity 2 reader access on the private Company.
			testUtils.AddDACActorRelationship{
				CollectionID:      1,
				DocID:             0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
			},
			// Identity 2 now has read access — linking should succeed.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(2),
				DocMap: map[string]any{
					"name":    "Employee",
					"salary":  50000,
					"company": testUtils.NewDocIndex(1, 0),
				},
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationAdd_RelationTarget_PublicDoc_NoIdentity_NoError asserts that any caller,
// including one without an identity, can link to a public (ACP-unregistered) document.
func TestACP_MutationAdd_RelationTarget_PublicDoc_NoIdentity_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			// No identity — the Company is public (not registered with ACP).
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Public Company",
					"capital": 100000
				}`,
			},
			// No-identity caller links to a public Company — should succeed.
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":    "Employee",
					"salary":  50000,
					"company": testUtils.NewDocIndex(1, 0),
				},
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationUpdate_RelationTarget_PrivateDoc_NoIdentity_Error asserts that a caller
// with no identity cannot update a relation field to point to a private document.
func TestACP_MutationUpdate_RelationTarget_PrivateDoc_NoIdentity_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			// Identity 1 creates a private Company (CollectionID 1, index 0).
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// Public Employee created without a company (CollectionID 0, index 0).
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Employee",
					"salary": 50000
				}`,
			},
			// No-identity caller tries to update the Employee's company to the private Company.
			&action.UpdateDoc{
				CollectionID:  0,
				DocID:         0,
				ExpectedError: "relation target document not found",
				DocMap: map[string]any{
					"_companyID": testUtils.NewDocIndex(1, 0),
				},
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestACP_MutationUpdate_RelationTarget_PrivateDoc_OwnerIdentity_NoError asserts that the
// owner of a private document can update a relation field to point to it.
func TestACP_MutationUpdate_RelationTarget_PrivateDoc_OwnerIdentity_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: append(
			policyAndSchemaSetup(),
			// Identity 1 creates a private Company (CollectionID 1, index 0).
			&action.AddDoc{
				CollectionID: 1,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Private Company",
					"capital": 200000
				}`,
			},
			// Identity 1 creates an Employee without a company (CollectionID 0, index 0).
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Employee",
					"salary": 50000
				}`,
			},
			// Identity 1 updates the Employee's company to their private Company — should succeed.
			&action.UpdateDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				DocID:        0,
				DocMap: map[string]any{
					"_companyID": testUtils.NewDocIndex(1, 0),
				},
			},
		),
	}
	testUtils.ExecuteTestCase(t, test)
}
