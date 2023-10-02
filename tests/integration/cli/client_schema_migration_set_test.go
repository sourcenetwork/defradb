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
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationSet_GivenEmptyArgs_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{"client", "schema", "migration", "set"})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing arguments. Required: src, dst, cfg")
}

func TestSchemaMigrationSet_GivenOneArg_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing arguments. Required: src, dst, cfg")
}

func TestSchemaMigrationSet_GivenTwoArgs_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing argument. Name: cfg")
}

func TestSchemaMigrationSet_GivenFourArgs_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", "cfg", "extraArg",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "too many arguments. Max: 3, Actual: 4")
}

func TestSchemaMigrationSet_GivenEmptySrcArg_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"", "bae", "path",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing argument. Name: src")
}

func TestSchemaMigrationSet_GivenEmptyDstArg_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae", "", "path",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing argument. Name: dst")
}

func TestSchemaMigrationSet_GivenEmptyCfgArg_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", "",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "missing argument. Name: cfg")
}

func TestSchemaMigrationSet_GivenInvalidCfgJsonObject_ShouldError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", "{--notvalidjson",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "invalid lens configuration: invalid character")
}

func TestSchemaMigrationSet_GivenEmptyCfgObject_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", "{}",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, "success")
}

func TestSchemaMigrationSet_GivenCfgWithNoLenses_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", `{"lenses": []}`,
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, "success")
}

func TestSchemaMigrationSet_GivenCfgWithNoLensesUppercase_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", `{"Lenses": []}`,
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, "success")
}

func TestSchemaMigrationSet_GivenCfgWithUnknownProp_ShouldError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", `{"NotAProp": []}`,
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "invalid lens configuration: json: unknown field")
}

func TestSchemaMigrationSet_GivenCfgWithUnknownPath_ShouldError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", `{"Lenses": [{"path":"notAPath"}]}`,
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "no such file or directory")
}

func TestSchemaMigrationSet_GivenCfgWithLenses_ShouldSucceedAndMigrateDoc(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", `type Users { name: String }`})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", `mutation { create_Users(data:"{\"name\":\"John\"}") { name } }`})
	assertContainsSubstring(t, stdout, `{"data":[{"name":"John"}]}`)

	stdout, _ = runDefraCommand(t, conf, []string{"client", "schema", "patch",
		`[{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }]`,
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bafkreibqw2l325up2tljc5oyjpjzftg4x7nhluzqoezrmz645jto6tnylu",
		"bafkreia56p6i6o3l4jijayiqd5eiijsypjjokbldaxnmqgeav6fe576hcy",
		fmt.Sprintf(`{"lenses": [{"path":"%s","arguments":{"dst":"verified","value":true}}]}`, lenses.SetDefaultModulePath),
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", "query { Users { name verified } }"})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, `{"data":[{"name":"John","verified":true}]}`)
}

func TestSchemaMigrationSet_GivenCfgWithLenseError_ShouldError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "schema", "add", `type Users { name: String }`})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", `mutation { create_Users(data:"{\"name\":\"John\"}") { name } }`})
	assertContainsSubstring(t, stdout, `{"data":[{"name":"John"}]}`)

	stdout, _ = runDefraCommand(t, conf, []string{"client", "schema", "patch",
		`[{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }]`,
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bafkreibqw2l325up2tljc5oyjpjzftg4x7nhluzqoezrmz645jto6tnylu",
		"bafkreia56p6i6o3l4jijayiqd5eiijsypjjokbldaxnmqgeav6fe576hcy",
		// Do not set lens parameters in order to generate error
		fmt.Sprintf(`{"lenses": [{"path":"%s"}]}`, lenses.SetDefaultModulePath),
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{"client", "query", "query { Users { name verified } }"})
	_ = stopDefra()

	// Error generated from within lens module lazily executing within the query
	assertContainsSubstring(t, stdout, "Parameters have not been set.")
}
