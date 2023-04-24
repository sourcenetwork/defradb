// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clitest

import (
	"testing"
)

func TestClientSchemaPatch(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	defer stopDefra()

	fname := schemaFileFixture(t, "schema.graphql", `
	type User {
		id: ID
		name: String
	}`)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "schema", "patch", `[{ "op": "add", "path": "/User/Schema/Fields/-", "value": {"Name": "address", "Kind": "String"} }]`})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", `query IntrospectionQuery { __type (name: "User") { fields { name } }}`})
	assertContainsSubstring(t, stdout, "address")
}

func TestClientSchemaPatch_InvalidJSONPatch(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	defer stopDefra()

	fname := schemaFileFixture(t, "schema.graphql", `
        type User {
            id: ID
            name: String
        }
    `)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "schema", "patch", `[{ "op": "invalidOp" }]`})
	assertContainsSubstring(t, stdout, "Internal Server Error")
}
