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
		Indexes: []IndexDescription{
			{
				Name:            "Book_title_ASC",
				ID:              1,
				Fields:          []IndexedFieldDescription{{Name: "title", Descending: false}},
				Kind:            IndexKindOrdered,
				KindDescription: &OrderedIndexDescription{Unique: false},
			},
			{
				Name:            "Book_pages_ASC",
				ID:              2,
				Fields:          []IndexedFieldDescription{{Name: "pages", Descending: false}},
				Kind:            IndexKindVector,
				KindDescription: &VectorIndexDescription{Dimensions: 128},
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

	// Index kinds render as their string form in display output.
	assert.Contains(t, out, `"Kind":"ordered"`)
	assert.Contains(t, out, `"Kind":"vector"`)

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

func TestCollectionVersion_StorageSerialization_EmitsNumeric(t *testing.T) {
	col := sampleDisplayCollectionVersion()

	// Persistence to storage uses json.Marshal directly. It must emit numeric Kind values
	// so storage footprint is not increased.
	rawStorage, err := json.Marshal(col)
	require.NoError(t, err)
	storageStr := string(rawStorage)

	assert.Contains(t, storageStr, `"Kind":0`)
	assert.Contains(t, storageStr, `"Kind":1`)
	assert.NotContains(t, storageStr, `"Kind":"ordered"`)
	assert.NotContains(t, storageStr, `"Kind":"vector"`)
}

func TestIndexDescription_Display(t *testing.T) {
	desc := IndexDescription{
		Name:            "Book_title_ASC",
		ID:              1,
		Fields:          []IndexedFieldDescription{{Name: "title"}},
		Kind:            IndexKindOrdered,
		KindDescription: &OrderedIndexDescription{Unique: false},
	}
	raw, err := desc.Display()
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"Kind":"ordered"`)
}
