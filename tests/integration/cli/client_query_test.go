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

func TestRequestSimple(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "query",
		"query IntrospectionQuery {__schema {queryType { name }}}",
	})
	nodeLog := stopDefra()

	assertContainsSubstring(t, stdout, "Query")
	assertNotContainsSubstring(t, nodeLog, "ERROR")
}

func TestRequestInvalidQuery(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "query", "{}}"})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, "Syntax Error")
}

func TestRequestWithErrorNoType(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	defer stopDefra()

	stdout, _ := runDefraCommand(t, conf, []string{"client", "query", "query { User { whatever } }"})

	assertContainsSubstring(t, stdout, "Cannot query field")
}

func TestRequestWithErrorNoField(t *testing.T) {
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

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", "query { User { nonexistent } }"})

	assertContainsSubstring(t, stdout, `Cannot query field \"nonexistent\"`)
}

func TestRequestQueryFromFile(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	defer stopDefra()

	fname := schemaFileFixture(t, "schema.graphql", `
		type User123 {
			XYZ: String
		}`)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fname})
	assertContainsSubstring(t, stdout, "success")

	fname = schemaFileFixture(t, "query.graphql", `
		query {
			__schema {
				types {
					name
					fields {
						name
						type {
							name
							kind
						}
					}
				}
			}
		}`)
	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", "-f", fname})

	assertContainsSubstring(t, stdout, "Query")

	// Check that the User type is correctly returned
	assertContainsSubstring(t, stdout, "User123")
	assertContainsSubstring(t, stdout, "XYZ")
}
