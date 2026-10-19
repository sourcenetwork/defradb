// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allCTypes are every CRDT type with a defined string representation.
var allCTypes = []CType{NONE_CRDT, LWW_REGISTER, OBJECT, COMPOSITE, PN_COUNTER, P_COUNTER}

func TestCTypeStringToFromIdentity(t *testing.T) {
	for _, ctype := range allCTypes {
		got, ok := CTypeFromString(ctype.String())
		require.True(t, ok, "String() %q should round-trip via CTypeFromString", ctype.String())
		assert.Equal(t, ctype, got)
	}
}

func TestCTypeFromString_Unknown(t *testing.T) {
	_, ok := CTypeFromString("not-a-crdt")
	assert.False(t, ok)
}

func TestCType_UnmarshalJSON_Numeric(t *testing.T) {
	// The numeric form is the persisted/patch form and must keep working unchanged.
	for _, ctype := range allCTypes {
		var got CType
		require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf("%d", byte(ctype))), &got))
		assert.Equal(t, ctype, got)
	}
}

func TestCType_UnmarshalJSON_String(t *testing.T) {
	// The string form is the display form emitted by collection describe.
	for _, ctype := range allCTypes {
		var got CType
		require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf("%q", ctype.String())), &got))
		assert.Equal(t, ctype, got)
	}
}

func TestCType_UnmarshalJSON_UnknownString(t *testing.T) {
	var got CType
	assert.Error(t, json.Unmarshal([]byte(`"not-a-crdt"`), &got))
}
