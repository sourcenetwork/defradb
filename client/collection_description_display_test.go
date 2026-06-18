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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleDisplayCollectionVersion holds a scalar, a scalar-array, and both relation kinds so
// the display rendering exercises every Kind encoding branch plus the Typ stringification.
func sampleDisplayCollectionVersion() CollectionVersion {
	return CollectionVersion{
		Name:         "Book",
		CollectionID: "bookid",
		IsActive:     true,
		Fields: []CollectionFieldDescription{
			{
				Name: "title",
				Kind: FieldKind_NILLABLE_STRING,
				Typ:  LWW_REGISTER,
			},
			{
				Name: "pages",
				Kind: FieldKind_INT_ARRAY,
				Typ:  LWW_REGISTER,
			},
			{
				Name: "author",
				Kind: NewCollectionKind("authorid", false),
				Typ:  NONE_CRDT,
			},
			{
				Name: "self",
				Kind: NewSelfKind("1", true),
				Typ:  NONE_CRDT,
			},
		},
	}
}

func TestCollectionVersionDisplay_RendersStrings(t *testing.T) {
	raw, err := sampleDisplayCollectionVersion().Display()
	require.NoError(t, err)
	out := string(raw)

	// Scalars and arrays render as their string form, Typ as the CRDT string.
	assert.Contains(t, out, `"Kind":"String"`)
	assert.Contains(t, out, `"Kind":"[Int!]"`)
	assert.Contains(t, out, `"Typ":"lww"`)
	assert.Contains(t, out, `"Typ":"none"`)

	// Relation kinds keep their object shape so they round-trip (not a bare string).
	assert.Contains(t, out, `"CollectionID":"authorid"`)
	assert.Contains(t, out, `"RelativeID":"1"`)

	// The numeric forms must NOT leak into the display output.
	assert.NotContains(t, out, `"Kind":11`)
	assert.NotContains(t, out, `"Typ":1`)
}

func TestCollectionVersionDisplay_RoundTrips(t *testing.T) {
	original := sampleDisplayCollectionVersion()

	raw, err := original.Display()
	require.NoError(t, err)

	// The display output unmarshals back into the original CollectionVersion (this is what
	// the HTTP client relies on when reading a describe response).
	var got CollectionVersion
	require.NoError(t, json.Unmarshal(raw, &got))
	assert.True(t, original.Equal(got), "Display() output must round-trip back into the original CollectionVersion")
}
