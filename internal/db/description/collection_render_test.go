// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

// sampleCollectionVersion holds a scalar, a scalar-array, and both relation kinds so the
// rendered form exercises every Kind encoding branch plus the Typ stringification.
func sampleCollectionVersion() client.CollectionVersion {
	return client.CollectionVersion{
		Name:         "Book",
		CollectionID: "bookid",
		IsActive:     true,
		Fields: []client.CollectionFieldDescription{
			{
				Name: "title",
				Kind: client.FieldKind_NILLABLE_STRING,
				Typ:  client.LWW_REGISTER,
			},
			{
				Name: "pages",
				Kind: client.FieldKind_INT_ARRAY,
				Typ:  client.LWW_REGISTER,
			},
			{
				Name: "author",
				Kind: client.NewCollectionKind("authorid", false),
				Typ:  client.NONE_CRDT,
			},
			{
				Name: "self",
				Kind: client.NewSelfKind("1", true),
				Typ:  client.NONE_CRDT,
			},
		},
	}
}

func TestRenderCollectionVersion_RendersStrings(t *testing.T) {
	rendered, err := description.RenderCollectionVersion(sampleCollectionVersion())
	require.NoError(t, err)

	b, err := json.Marshal(rendered)
	require.NoError(t, err)
	out := string(b)

	// Scalars and arrays render as their string form, Typ as the CRDT string.
	assert.Contains(t, out, `"Kind":"String"`)
	assert.Contains(t, out, `"Kind":"[Int!]"`)
	assert.Contains(t, out, `"Typ":"lww"`)
	assert.Contains(t, out, `"Typ":"none"`)

	// Relation kinds keep their object shape so they round-trip (not a bare string).
	assert.Contains(t, out, `"CollectionID":"authorid"`)
	assert.Contains(t, out, `"RelativeID":"1"`)

	// The numeric forms must NOT leak into the rendered output.
	assert.NotContains(t, out, `"Kind":11`)
	assert.NotContains(t, out, `"Typ":1`)
}

func TestRenderCollectionVersions_RoundTrips(t *testing.T) {
	original := sampleCollectionVersion()

	rendered, err := description.RenderCollectionVersions([]client.CollectionVersion{original})
	require.NoError(t, err)

	b, err := json.Marshal(rendered)
	require.NoError(t, err)

	// The HTTP client unmarshals the describe response back into []client.CollectionVersion.
	var got []client.CollectionVersion
	require.NoError(t, json.Unmarshal(b, &got))

	require.Len(t, got, 1)
	assert.True(t, original.Equal(got[0]), "rendered form must round-trip back into the original CollectionVersion")
}
