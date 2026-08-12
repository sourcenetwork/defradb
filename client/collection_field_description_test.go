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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A generic any-decode of a JSON number always yields float64, and of a JSON
// string always yields string, regardless of what Kind says the field actually
// holds. However, our unmarshalling should respect the concrete Go types that were
// initially passed. These are tests of that correct unmarshalling behavior.

func TestCollectionFieldDescription_UnmarshalJSON_IntDefault_DecodesAsInt32(t *testing.T) {
	raw := []byte(`{"Name":"age","Kind":4,"DefaultValue":5}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, int32(5), f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_Float32Default_DecodesAsFloat32(t *testing.T) {
	raw := []byte(`{"Name":"score","Kind":8,"DefaultValue":1.5}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, float32(1.5), f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_Float64Default_DecodesAsFloat64(t *testing.T) {
	raw := []byte(`{"Name":"score","Kind":6,"DefaultValue":1.5}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, float64(1.5), f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_DateTimeDefault_DecodesAsTimeTime(t *testing.T) {
	raw := []byte(`{"Name":"createdAt","Kind":10,"DefaultValue":"2024-01-02T03:04:05Z"}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	expected, err := time.Parse(time.RFC3339, "2024-01-02T03:04:05Z")
	require.NoError(t, err)
	assert.Equal(t, expected, f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_DateTimeUTCNowSentinel_PreservedAsString(t *testing.T) {
	raw := []byte(`{"Name":"createdAt","Kind":10,"DefaultValue":"UTC_NOW"}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, UTCNOW, f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_BoolDefault_DecodesAsBool(t *testing.T) {
	raw := []byte(`{"Name":"active","Kind":2,"DefaultValue":true}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, true, f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_StringDefault_DecodesAsString(t *testing.T) {
	raw := []byte(`{"Name":"name","Kind":11,"DefaultValue":"hello"}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Equal(t, "hello", f.DefaultValue)
}

func TestCollectionFieldDescription_UnmarshalJSON_NoDefault_IsNil(t *testing.T) {
	raw := []byte(`{"Name":"name","Kind":11}`)

	var f CollectionFieldDescription
	err := json.Unmarshal(raw, &f)
	require.NoError(t, err)

	assert.Nil(t, f.DefaultValue)
}