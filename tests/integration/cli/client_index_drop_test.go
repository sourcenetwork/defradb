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

func TestIndexDrop_IfNoArgs_ShowUsage(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"client", "index", "drop"})
	assertContainsSubstring(t, stderr, "Usage")
}

const userColIndexOnNameFieldName = "users_name_index"

func createIndexOnName(t *testing.T, conf DefraNodeConfig) {
	createIndexOnField(t, conf, "User", "name", userColIndexOnNameFieldName)
}

func createIndexOnField(t *testing.T, conf DefraNodeConfig, colName, fieldName, indexName string) {
	runDefraCommand(t, conf, []string{
		"client", "index", "create",
		"--collection", colName,
		"--fields", fieldName,
		"--name", indexName,
	})
}

func TestIndexDrop_IfNoNameArg_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)
	createIndexOnName(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "index", "drop",
		"--collection", "User",
	})
	stopDefra()

	assertContainsSubstring(t, stderr, "missing argument")
}

func TestIndexDrop_IfNoCollectionArg_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)
	createIndexOnName(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "index", "drop",
		"--name", "users_name_index",
	})
	stopDefra()

	assertContainsSubstring(t, stderr, "missing argument")
}

func TestIndexDrop_IfCollectionWithIndexExists_ShouldDropIndex(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)
	createIndexOnName(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "drop",
		"--collection", "User",
		"--name", "users_name_index",
	})
	nodeLog := stopDefra()

	jsonResponse := `{"data":{"result":"success"}}`
	assertContainsSubstring(t, stdout, jsonResponse)
	assertNotContainsSubstring(t, stdout, "errors")
	assertNotContainsSubstring(t, nodeLog, "errors")
}

func TestIndexDrop_IfCollectionDoesNotExist_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "drop",
		"--collection", "User",
		"--name", "users_name_index",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "errors")
}

func TestIndexDrop_IfInternalError_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "drop",
		"--collection", "User",
		"--name", "users_name_index",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "errors")
}
