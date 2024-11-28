// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONEncodingAndDecoding_ShouldEncodeAndDecodeBack(t *testing.T) {
	jsonMap := map[string]any{
		"str":  "value",
		"num":  123.5,
		"bool": true,
		"null": nil,
		"obj": map[string]any{
			"obj_str":  "obj_val",
			"obj_num":  42,
			"obj_bool": false,
			"obj_null": nil,
			"obj_obj": map[string]any{
				"obj_obj_str": "obj_obj_val",
			},
			"obj_arr": []any{"obj_arr_val", 100},
		},
		"arr": []any{
			"arr_val",
			23,
			false,
			nil,
			map[string]any{
				"arr_obj": "arr_obj_val",
			},
			[]any{"arr_arr_val", 1000},
		},
	}

	testJSON, err := client.NewJSON(jsonMap)
	assert.NoError(t, err)

	pathMap := make(map[string][]client.JSON)

	err = client.TraverseJSON(testJSON, func(value client.JSON) error {
		p := strings.Join(value.GetPath(), "/")
		jsons := pathMap[p]
		jsons = append(jsons, value)
		pathMap[p] = jsons
		return nil
	}, client.TraverseJSONOnlyLeaves(), client.TraverseJSONVisitArrayElements())
	assert.NoError(t, err)

	for path, jsons := range pathMap {
		for i, value := range jsons {
			for _, ascending := range []bool{true, false} {
				t.Run(fmt.Sprintf("Path %s, index: %d, ascending: %v", path, i, ascending), func(t *testing.T) {
					var encoded []byte
					if ascending {
						encoded = EncodeJSONAscending(nil, value)
					} else {
						encoded = EncodeJSONDescending(nil, value)
					}

					var remaining []byte
					var decoded client.JSON
					var err error

					if ascending {
						remaining, decoded, err = DecodeJSONAscending(encoded)
					} else {
						remaining, decoded, err = DecodeJSONDescending(encoded)
					}

					require.NoError(t, err)
					assert.Empty(t, remaining)
					assert.Equal(t, value.Value(), decoded.Value())
				})
			}
		}
	}
}

func TestJSONDecoding_MalformedData(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		ascending   bool
		expectedErr string
	}{
		{
			name:      "malformed json path",
			input:     []byte{jsonMarker, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			ascending: true,
		},
		{
			name:      "malformed json num",
			input:     []byte{jsonMarker, ascendingBytesEscapes.escapedTerm, floatPos, 0xFF, 0xFF, 0xFF},
			ascending: true,
		},
		{
			name:      "malformed json num",
			input:     []byte{jsonMarker, ascendingBytesEscapes.escapedTerm, floatPos, 0xFF, 0xFF, 0xFF},
			ascending: false,
		},
		{
			name:      "malformed json num",
			input:     []byte{jsonMarker, ascendingBytesEscapes.escapedTerm, bytesMarker, 0xFF, 0xFF, 0xFF},
			ascending: true,
		},
		{
			name:      "malformed json num",
			input:     []byte{jsonMarker, ascendingBytesEscapes.escapedTerm, bytesDescMarker, 0xFF, 0xFF, 0xFF},
			ascending: false,
		},
		{
			name:      "wrong type marker",
			input:     []byte{jsonMarker, ascendingBytesEscapes.escapedTerm, timeMarker, 0xFF, 0xFF, 0xFF},
			ascending: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			if tt.ascending {
				_, _, err = DecodeJSONAscending(tt.input)
			} else {
				_, _, err = DecodeJSONDescending(tt.input)
			}

			assert.Error(t, err)
		})
	}
}
