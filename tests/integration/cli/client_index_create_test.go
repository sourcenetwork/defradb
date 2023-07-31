// Copyright 2023 Democratized Data Foundation
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

func createUserCollection(t *testing.T, conf DefraNodeConfig) {
	createCollection(t, conf, `type User { name: String }`)
}

func createCollection(t *testing.T, conf DefraNodeConfig, colSchema string) {
	fileName := schemaFileFixture(t, "schema.graphql", colSchema)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", "-f", fileName})
	assertContainsSubstring(t, stdout, "success")
}

func TestIndex_IfNoArgs_ShowUsage(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "index"})
	assertContainsSubstring(t, stdout, "Usage:")
}

func TestIndexCreate_IfNoArgs_ShowUsage(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"client", "index", "create"})
	assertContainsSubstring(t, stderr, "Usage")
}

func TestIndexCreate_IfNoFieldsArg_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "index", "create",
		"--collection", "User",
	})
	stopDefra()

	assertContainsSubstring(t, stderr, "missing argument")
}

func TestIndexCreate_IfNoCollectionArg_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "index", "create",
		"--fields", "Name",
	})
	stopDefra()

	assertContainsSubstring(t, stderr, "missing argument")
}

func TestIndexCreate_IfCollectionExists_ShouldCreateIndex(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "create",
		"--collection", "User",
		"--fields", "name",
		"--name", "users_name_index",
	})
	nodeLog := stopDefra()

	jsonResponse := `{"data":{"index":{"Name":"users_name_index","ID":1,"Fields":[{"Name":"name","Direction":"ASC"}]}}}`
	assertContainsSubstring(t, stdout, jsonResponse)
	assertNotContainsSubstring(t, stdout, "errors")
	assertNotContainsSubstring(t, nodeLog, "errors")
}

func TestIndexCreate_IfInternalError_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "create",
		"--collection", "User",
		"--fields", "Name",
		"--name", "users_name_index",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "errors")
}
