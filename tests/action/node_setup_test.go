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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/tests/state"
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
	"NodeACP":          fieldFlagged,

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

// testOwner is the identity a node under access control is started with.
var testOwner = immutable.Some(state.Identity{
	Kind:     state.ClientIdentityType,
	Selector: "1",
})

// newFlagsTestState builds the minimum state externalNodeFlags reads.
func newFlagsTestState(t *testing.T, keyType crypto.KeyType) *state.State {
	return &state.State{
		T:          t,
		DbType:     DefraIMType,
		Identities: map[state.Identity]*state.IdentityHolder{},
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testOwner.Value(): keyType,
		},
	}
}

func TestExternalNodeFlags_NAC(t *testing.T) {
	nacOn := immutable.Some(options.NodeACPOptions{IsEnabled: true})

	tests := []struct {
		name            string
		nodeACP         immutable.Option[options.NodeACPOptions]
		identity        immutable.Option[state.Identity]
		keyType         crypto.KeyType
		wantFlags       []string
		wantNoFlags     []string
		wantUnsupported string
	}{
		{
			name:        "off by default",
			identity:    testOwner,
			keyType:     crypto.KeyTypeSecp256k1,
			wantNoFlags: []string{"--node-acp-enable", "--identity"},
		},
		{
			name:        "set but disabled stays off",
			nodeACP:     immutable.Some(options.NodeACPOptions{IsEnabled: false}),
			identity:    testOwner,
			keyType:     crypto.KeyTypeSecp256k1,
			wantNoFlags: []string{"--node-acp-enable", "--identity"},
		},
		{
			name:      "enabled asks the node to turn it on",
			nodeACP:   nacOn,
			identity:  testOwner,
			keyType:   crypto.KeyTypeSecp256k1,
			wantFlags: []string{"--node-acp-enable", "--identity"},
		},
		{
			// The node refuses to start access control with no one to own it.
			name:            "no identity to own it",
			nodeACP:         nacOn,
			identity:        immutable.None[state.Identity](),
			keyType:         crypto.KeyTypeSecp256k1,
			wantNoFlags:     []string{"--node-acp-enable"},
			wantUnsupported: "no private key",
		},
		{
			// The key is sent as bare hex and read back as secp256k1, so another
			// type would silently give access control to a different owner.
			name:            "owner key the node cannot read",
			nodeACP:         nacOn,
			identity:        testOwner,
			keyType:         crypto.KeyTypeEd25519,
			wantNoFlags:     []string{"--node-acp-enable"},
			wantUnsupported: "secp256k1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newFlagsTestState(t, test.keyType)

			flags, unsupported := externalNodeFlags(s, test.identity,
				NodeSetupConfig{NodeACP: test.nodeACP})

			joinedFlags := strings.Join(flags, " ")
			for _, want := range test.wantFlags {
				assert.Contains(t, joinedFlags, want)
			}
			for _, notWant := range test.wantNoFlags {
				assert.NotContains(t, joinedFlags, notWant)
			}

			joinedUnsupported := strings.Join(unsupported, " ")
			if test.wantUnsupported == "" {
				assert.NotContains(t, joinedUnsupported, "node access control")
			} else {
				assert.Contains(t, joinedUnsupported, test.wantUnsupported)
			}
		})
	}
}
