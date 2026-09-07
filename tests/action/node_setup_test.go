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
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupFieldHandling records how externalNodeFlags treats each [NodeSetupConfig]
// field, so a new one cannot be added without saying which it is.
//
// Node access control was once dropped there, and the tests asserting it passed
// against a node that never enforced it.
type setupFieldHandling int

const (
	// fieldFlagged is read and turned into a flag, or reported as unsupported so
	// the test skips.
	fieldFlagged setupFieldHandling = iota
	// fieldDropped is not read at all. The external node runs the default, so a
	// test relying on the field is not running what it asks for.
	fieldDropped
)

var setupFields = map[string]setupFieldHandling{
	"EnableSigning":    fieldFlagged,
	"HTTP":             fieldFlagged,
	"BadgerEncryption": fieldFlagged,
	"EnableNAC":        fieldFlagged,
	"NACOwner":         fieldFlagged,

	// The external node keeps its store under the rootdir the wrapper owns, so it
	// cannot reopen one the harness chose.
	"DatabaseDir": fieldDropped,
	// Both are only read when a lens is configured, which no cross-version test
	// does yet.
	"LensRuntime":  fieldDropped,
	"LensPoolSize": fieldDropped,
	// Both belong to remote document ACP, which is already reported as
	// unsupported before either is looked at.
	"IsDocumentACPTest": fieldDropped,
	"VeraImage":         fieldDropped,
}

func TestNodeSetupConfig_EveryFieldIsAccountedForByExternalNodeFlags(t *testing.T) {
	cfgType := reflect.TypeOf(NodeSetupConfig{})

	var actual []string
	for i := range cfgType.NumField() {
		actual = append(actual, cfgType.Field(i).Name)
	}

	var known []string
	for name := range setupFields {
		known = append(known, name)
	}

	assert.ElementsMatch(t, known, actual,
		"NodeSetupConfig changed. Give the new field a flag in externalNodeFlags, or "+
			"report it as unsupported so the test skips, then record it in setupFields.")
}

// TestExternalNodeFlags_ReadsEveryFlaggedField fails if a field recorded as
// flagged stops being read, which would drop it silently.
func TestExternalNodeFlags_ReadsEveryFlaggedField(t *testing.T) {
	source := readNodeSetupSource(t)

	for name, handling := range setupFields {
		if handling != fieldFlagged {
			continue
		}
		assert.True(t, strings.Contains(source, "cfg."+name),
			"%s is recorded as flagged in setupFields, but externalNodeFlags no longer "+
				"reads it. Either read it again or record it as dropped.", name)
	}
}

// readNodeSetupSource returns the body of externalNodeFlags.
func readNodeSetupSource(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile("node_setup.go")
	require.NoError(t, err)

	start := strings.Index(string(source), "func externalNodeFlags(")
	require.NotEqual(t, -1, start, "externalNodeFlags was renamed or moved")

	body := string(source)[start:]
	end := strings.Index(body, "\n}\n")
	require.NotEqual(t, -1, end, "could not find the end of externalNodeFlags")

	return body[:end]
}
