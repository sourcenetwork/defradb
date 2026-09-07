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

//go:build !js

package action

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// knownSetupFields are the [NodeSetupConfig] fields externalNodeFlags has been
// checked against, each either a flag or named as unsupported.
//
// Node access control was once dropped there, and the tests asserting it passed
// against a node that never enforced it.
var knownSetupFields = []string{
	"EnableSigning",
	"HTTP",
	"IsDocumentACPTest",
	"VeraImage",
	"DatabaseDir",
	"BadgerEncryption",
	"LensRuntime",
	"LensPoolSize",
	"EnableNAC",
	"NACOwner",
}

func TestNodeSetupConfig_EveryFieldIsHandledByExternalNodeFlags(t *testing.T) {
	var actual []string
	for i := range reflect.TypeOf(NodeSetupConfig{}).NumField() {
		actual = append(actual, reflect.TypeOf(NodeSetupConfig{}).Field(i).Name)
	}

	assert.ElementsMatch(t, knownSetupFields, actual,
		"NodeSetupConfig changed. Give the new field a flag in externalNodeFlags, or "+
			"add it to the unsupported list so the test skips, then list it here.")
}
