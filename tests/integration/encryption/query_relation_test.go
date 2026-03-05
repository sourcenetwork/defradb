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

package encryption

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryption_WithEncryptionOnBothRelations_ShouldFetchDecrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `query {
					User {
						name
						devices {
							model
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnPrimaryRelations_ShouldFetchDecrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `query {
					User {
						name
						devices {
							model
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnSecondaryRelations_ShouldFetchDecrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
						devices {
							model
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
