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

	"github.com/stretchr/testify/require"
)

// GQLResult does not use the default struct marshalling. It copies itself field by
// field through a private mirror, so a field can exist on the type and still be missing
// from the JSON. These tests catch that.
//
// It matters because the Go client never serializes anything. A half applied change
// passes there and fails on every other client.

func TestGQLResultMarshal_WithWarning_RoundTrips(t *testing.T) {
	input := GQLResult{
		Data: map[string]any{"Users": []any{}},
		Extensions: &GQLExtensions{
			Warnings: []GQLWarning{
				{
					Code:    "test_warning",
					Message: "something worth knowing happened",
					Detail:  map[string]any{"requested": 10, "returned": 6},
				},
			},
		},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var output GQLResult
	err = json.Unmarshal(data, &output)
	require.NoError(t, err)

	require.NotNil(t, output.Extensions)
	require.Len(t, output.Extensions.Warnings, 1)

	warning := output.Extensions.Warnings[0]
	require.Equal(t, "test_warning", warning.Code)
	require.Equal(t, "something worth knowing happened", warning.Message)

	// UnmarshalJSON calls dec.UseNumber, so numbers come back as json.Number, not
	// float64. Same as everything under `data`.
	require.Equal(t, json.Number("10"), warning.Detail["requested"])
	require.Equal(t, json.Number("6"), warning.Detail["returned"])
}

func TestGQLResultMarshal_WithMultipleWarnings_PreservesOrder(t *testing.T) {
	input := GQLResult{
		Extensions: &GQLExtensions{
			Warnings: []GQLWarning{
				{Code: "first", Message: "one"},
				{Code: "second", Message: "two"},
			},
		},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var output GQLResult
	err = json.Unmarshal(data, &output)
	require.NoError(t, err)

	require.Len(t, output.Extensions.Warnings, 2)
	require.Equal(t, "first", output.Extensions.Warnings[0].Code)
	require.Equal(t, "second", output.Extensions.Warnings[1].Code)
}

func TestGQLResultMarshal_WithoutExtensions_OmitsField(t *testing.T) {
	input := GQLResult{Data: map[string]any{"Users": []any{}}}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	require.NotContains(t, raw, "extensions")
}

func TestGQLResultMarshal_WithEmptyExtensions_OmitsField(t *testing.T) {
	// An empty value must not be sent as `"extensions":{}`. Otherwise every response
	// changes shape as soon as anything allocates an accumulator.
	input := GQLResult{
		Data:       map[string]any{"Users": []any{}},
		Extensions: &GQLExtensions{},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	require.NotContains(t, raw, "extensions")
}

func TestGQLResultUnmarshal_WithUnknownExtensionKey_IsIgnored(t *testing.T) {
	// An older client must ignore an entry it does not know about instead of failing
	// the whole response.
	data := []byte(`{
		"data": null,
		"extensions": {
			"warnings": [{"code": "known", "message": "hi", "unknownField": 1}],
			"unknownKey": {"anything": true}
		}
	}`)

	var output GQLResult
	err := json.Unmarshal(data, &output)
	require.NoError(t, err)

	require.NotNil(t, output.Extensions)
	require.Len(t, output.Extensions.Warnings, 1)
	require.Equal(t, "known", output.Extensions.Warnings[0].Code)
}

func TestGQLResultUnmarshal_WithoutExtensions_LeavesNil(t *testing.T) {
	var output GQLResult
	err := json.Unmarshal([]byte(`{"data": null}`), &output)
	require.NoError(t, err)

	require.Nil(t, output.Extensions)
	require.True(t, output.Extensions.IsEmpty())
}

// The empty non nil slice is the case a field by field check gets wrong: the slice is
// not the zero value, but it still encodes to nothing.
func TestGQLExtensionsIsEmpty_WithEmptyWarningSlice_IsEmpty(t *testing.T) {
	extensions := &GQLExtensions{Warnings: []GQLWarning{}}

	require.True(t, extensions.IsEmpty())
}

func TestGQLExtensionsIsEmpty_WithNilReceiver_IsEmpty(t *testing.T) {
	var extensions *GQLExtensions

	require.True(t, extensions.IsEmpty())
}

func TestGQLExtensionsIsEmpty_WithAWarning_IsNotEmpty(t *testing.T) {
	extensions := &GQLExtensions{Warnings: []GQLWarning{{Code: "test_warning"}}}

	require.False(t, extensions.IsEmpty())
}

// A peer may send an empty extensions object. Callers are told the field is nil when
// there is nothing to report, so decoding must make that true rather than hand back a
// non nil value with nothing in it.
func TestGQLResultUnmarshal_WithEmptyExtensions_LeavesNil(t *testing.T) {
	var output GQLResult
	err := json.Unmarshal([]byte(`{"data": null, "extensions": {}}`), &output)
	require.NoError(t, err)

	require.Nil(t, output.Extensions)
}

func TestGQLResultUnmarshal_WithOnlyUnknownExtensionKeys_LeavesNil(t *testing.T) {
	var output GQLResult
	err := json.Unmarshal([]byte(`{"data": null, "extensions": {"unknownKey": 1}}`), &output)
	require.NoError(t, err)

	require.Nil(t, output.Extensions)
}
