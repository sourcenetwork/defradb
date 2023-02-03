// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
