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

package predefined

// DocsList is a list of document structures that might nest other documents to be replicated
// by a document generator.
//
//	gen.DocsList{
//		ColName: "User",
//		Docs: []map[string]any{
//			{
//				"name":     "Shahzad",
//				"age":      20,
//				"devices": []map[string]any{
//					{
//						"model": "iPhone Xs",
//					},
//				},
//			},
//		},
type DocsList struct {
	// ColName is the name of the collection that the documents in Docs belong to.
	ColName string
	// Docs is a list of documents to be replicated.
	Docs []map[string]any
}
