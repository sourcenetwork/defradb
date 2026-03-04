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

package fixtures

var (
	gTypeToGQLType = map[string]string{
		"int":     "Int",
		"string":  "String",
		"float64": "Float",
		"float32": "Float",
		"bool":    "Boolean",
	}
)

type User struct {
	Name     string `faker:"name"`
	Age      int
	Points   float32 `faker:"amount"`
	Verified bool
}
