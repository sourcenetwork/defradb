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

func TestSchemaMigrationGet_GivenOneArg_ShouldReturnError(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	_, stderr := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "get",
		"notAnArg",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stderr, "too many arguments. Max: 0, Actual: 1")
}

func TestSchemaMigrationGet_GivenNoMigrations_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "get",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout, `{"data":{"configuration":[]}}`)
}

func TestSchemaMigrationGet_GivenEmptyMigrationObj_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", "{}",
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "get",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout,
		`{"data":{"configuration":[{"SourceSchemaVersionID":"bae123","DestinationSchemaVersionID":"bae456","Lenses":null}]}}`,
	)
}

func TestSchemaMigrationGet_GivenEmptyMigration_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456", `{"lenses": []}`,
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "get",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout,
		`{"data":{"configuration":[{"SourceSchemaVersionID":"bae123","DestinationSchemaVersionID":"bae456","Lenses":[]}]}}`,
	)
}

func TestSchemaMigrationGet_GivenMigration_ShouldSucceed(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "set",
		"bae123", "bae456",
		fmt.Sprintf(`{"lenses": [{"path":"%s","arguments":{"dst":"verified","value":true}}]}`, lenses.SetDefaultModulePath),
	})
	assertContainsSubstring(t, stdout, "success")

	stdout, _ = runDefraCommand(t, conf, []string{
		"client", "schema", "migration", "get",
	})
	_ = stopDefra()

	assertContainsSubstring(t, stdout,
		`{"data":{"configuration":[{"SourceSchemaVersionID":"bae123","DestinationSchemaVersionID":"bae456","Lenses":[`+
			fmt.Sprintf(
				`{"Path":"%s",`,
				lenses.SetDefaultModulePath,
			)+
			`"Inverse":false,"Arguments":{"dst":"verified","value":true}}`+
			`]}]}}`,
	)
}
