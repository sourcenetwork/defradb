// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceID_WithEmpty_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/VersionID",
							"value": ""
						}
					]
				`,
				ExpectedError: "collections cannot be deleted.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithExisting_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
				Schema: `
					type Books {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/VersionID",
							"value": "fdsahasgsag"
						}
					]
				`,
				ExpectedError: "adding collections via patch is not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithExistingSameRoot_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/VersionID",
							"value": "bafkreialnju2rez4t3quvpobf3463eai3lo64vdrdhdmunz7yy7sv3f5ce"
						},
						{
							"op": "replace",
							"path": "/bafkreialnju2rez4t3quvpobf3463eai3lo64vdrdhdmunz7yy7sv3f5ce/VersionID",
							"value": "bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai"
						}
					]
				`,
				ExpectedError: "collection sources cannot be added or removed.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithExistingDifferentRoot_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
				Schema: `
					type Dogs {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/VersionID",
							"value": "bafkreibifvyfr6qvb6wx4v4cogvcdksb3v7vniaon7hdzzqb62cotpmlc4"
						},
						{
							"op": "replace",
							"path": "/bafkreibifvyfr6qvb6wx4v4cogvcdksb3v7vniaon7hdzzqb62cotpmlc4/VersionID",
							"value": "bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai"
						}
					]
				`,
				ExpectedError: "collection ID cannot be mutated.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithNew_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/VersionID",
							"value": "bafkreibifvyfr6qvb6wx4v4cogvcdksb3v7vniaon7hdzzqb62cotpmlc4"
						}
					]
				`,
				ExpectedError: "adding collections via patch is not supported.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
