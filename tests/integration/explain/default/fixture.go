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

package test_explain_default

type dataMap = map[string]any

var basicPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"scanNode": dataMap{},
					},
				},
			},
		},
	},
}

var emptyChildSelectsAttributeForAuthor = dataMap{
	"collectionName": "Author",
	"docID":          nil,
	"filter":         nil,
	"groupBy":        nil,
	"limit":          nil,
	"orderBy":        nil,
}
