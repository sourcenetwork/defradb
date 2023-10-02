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

	"github.com/stretchr/testify/assert"
)

func TestAddSchemaFromFile(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	fname := schemaFileFixture(t, "schema.graphql", `
	type User {
		id: ID
		name: String
	}`)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname})

	nodeLog := stopDefra()

	jsonReponse := `{"data":{"collections":[{"name":"User","id":"bafkreifxwnqwcg3uqqr3iydebnmeadmjxg722qauocdtjbusinjtzja7py","version_id":"bafkreifxwnqwcg3uqqr3iydebnmeadmjxg722qauocdtjbusinjtzja7py"}],"result":"success"}}`
	assert.Contains(t, stdout, jsonReponse)
	assertNotContainsSubstring(t, nodeLog, "ERROR")
}

func TestAddSchemaWithDuplicateType(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	fname1 := schemaFileFixture(t, "schema1.graphql", `type Post { id: ID title: String }`)
	fname2 := schemaFileFixture(t, "schema2.graphql", `type Post { id: ID author: String }`)

	stdout1, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname1})
	stdout2, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname2})

	_ = stopDefra()

	jsonReponse := `{"data":{"collections":[{"name":"Post","id":"bafkreibamgkyo3juvgx2b3ice4tjldcuxiibwo32kq22vfuyvzzgg7kfga","version_id":"bafkreibamgkyo3juvgx2b3ice4tjldcuxiibwo32kq22vfuyvzzgg7kfga"}],"result":"success"}}`
	assertContainsSubstring(t, stdout1, jsonReponse)
	assertContainsSubstring(t, stdout2, `schema type already exists. Name: Post`)
}
