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
	"testing"

	"github.com/stretchr/testify/require"
)

// MarshalFieldKindToJSON must round-trip back through parseFieldKind. Mapped scalar/array
// kinds go via their string name; an unmapped enum (whose String() is just the number) must
// be emitted numerically, otherwise parseFieldKind reads it back as a NamedKind.
func TestMarshalFieldKindToJSON_RoundTrips(t *testing.T) {
	kinds := []FieldKind{
		FieldKind_NILLABLE_STRING,
		FieldKind_STRING_ARRAY,
		FieldKind_DocID,
		ScalarKind(0),  // FieldKind_None - not in the string mapping
		ScalarKind(16), // deprecated/unused slot - not in the string mapping
		NewCollectionKind("somecollectionid", true),
		NewSelfKind("1", false),
	}

	for _, kind := range kinds {
		b, err := MarshalFieldKindToJSON(kind)
		require.NoError(t, err)

		got, err := parseFieldKind(b)
		require.NoError(t, err)
		require.Equalf(t, kind, got, "kind %T(%v) did not round-trip (json=%s)", kind, kind, string(b))
	}
}
