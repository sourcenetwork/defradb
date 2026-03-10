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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceID_WithEmpty_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/VersionID",
							"value": ""
						}
					]
				`,
				ExpectedError: "collection ID cannot be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithExisting_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VersionID",
							"value": "invalid cid"
						}
					]
				`,
				ExpectedError: "invalid cid: selected encoding not supported. VersionID: invalid cid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithExistingSameRoot_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/VersionID",
							"value": "bafyreieqhzanpek5ssb7ofi3qelbvl2nwh6s7x3w2mlzbcnqaqol3elltq"
						},
						{
							"op": "replace",
							"path": "/bafyreieqhzanpek5ssb7ofi3qelbvl2nwh6s7x3w2mlzbcnqaqol3elltq/VersionID",
							"value": "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"
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
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Dogs {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/VersionID",
							"value": "bafyreiguy7x6zs57dgbpuduiacckubvkbgi6bo2oaytu5dlthr2hsmawxu"
						},
						{
							"op": "replace",
							"path": "/bafyreiguy7x6zs57dgbpuduiacckubvkbgi6bo2oaytu5dlthr2hsmawxu/VersionID",
							"value": "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"
						}
					]
				`,
				ExpectedError: "collection source must belong to host collection.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceID_WithNew_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VersionID",
							"value": "bafkreibifvyfr6qvb6wx4v4cogvcdksb3v7vniaon7hdzzqb62cotpmlc4"
						}
					]
				`,
				ExpectedError: "unknown CID, collection ids cannot be manually defined",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
