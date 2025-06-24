// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestIndexDataStoreKey_PrefixEnd(t *testing.T) {
	tests := []struct {
		name            string
		key             IndexDataStoreKey
		wantGreaterThan bool // Should PrefixEnd be > original key bytes
	}{
		{
			name:            "empty key",
			key:             IndexDataStoreKey{},
			wantGreaterThan: false, // Empty key can't be incremented
		},
		{
			name: "collection only",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
			},
			wantGreaterThan: true,
		},
		{
			name: "collection and index",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
			},
			wantGreaterThan: true,
		},
		{
			name: "with single field",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalInt(42), Descending: false},
				},
			},
			wantGreaterThan: true,
		},
		{
			name: "with multiple fields",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("hello"), Descending: false},
					{Value: client.NewNormalInt(42), Descending: false},
				},
			},
			wantGreaterThan: true,
		},
		{
			name: "with descending field",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalInt(42), Descending: true},
				},
			},
			wantGreaterThan: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyBytes := tt.key.Bytes()
			endBytes := tt.key.PrefixEnd()

			if tt.wantGreaterThan {
				assert.True(t, string(endBytes) > string(keyBytes),
					"PrefixEnd should be greater than original key")
			} else {
				assert.Equal(t, keyBytes, endBytes,
					"PrefixEnd should equal original for empty key")
			}
		})
	}
}

func TestIndexDataStoreKey_PrefixEnd_Ordering(t *testing.T) {
	tests := []struct {
		name        string
		prefixKey   IndexDataStoreKey
		testKeys    []IndexDataStoreKey
		shouldMatch []bool // which test keys should be included
	}{
		{
			name: "single field prefix",
			prefixKey: IndexDataStoreKey{
				CollectionShortID: 1,
				IndexID:           2,
				Fields: []IndexedField{
					{Value: client.NewNormalInt(25), Descending: false},
				},
			},
			testKeys: []IndexDataStoreKey{
				{
					CollectionShortID: 1,
					IndexID:           2,
					Fields: []IndexedField{
						{Value: client.NewNormalInt(25), Descending: false},
						{Value: client.NewNormalString("Alice"), Descending: false},
					},
				},
				{
					CollectionShortID: 1,
					IndexID:           2,
					Fields: []IndexedField{
						{Value: client.NewNormalInt(25), Descending: false},
						{Value: client.NewNormalString("Bob"), Descending: false},
					},
				},
				{
					CollectionShortID: 1,
					IndexID:           2,
					Fields: []IndexedField{
						{Value: client.NewNormalInt(26), Descending: false},
					},
				},
			},
			shouldMatch: []bool{true, true, false},
		},
		{
			name: "collection prefix",
			prefixKey: IndexDataStoreKey{
				CollectionShortID: 10,
			},
			testKeys: []IndexDataStoreKey{
				{
					CollectionShortID: 10,
					IndexID:           1,
				},
				{
					CollectionShortID: 10,
					IndexID:           2,
				},
				{
					CollectionShortID: 11,
					IndexID:           1,
				},
			},
			shouldMatch: []bool{true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefixBytes := tt.prefixKey.Bytes()
			endBytes := tt.prefixKey.PrefixEnd()

			for i, testKey := range tt.testKeys {
				testBytes := testKey.Bytes()

				// Check if test key is >= prefix and < end
				afterPrefix := string(testBytes) >= string(prefixBytes)
				beforeEnd := string(testBytes) < string(endBytes)
				inRange := afterPrefix && beforeEnd

				assert.Equal(t, tt.shouldMatch[i], inRange,
					"Key %d: expected %v but got %v for key %v",
					i, tt.shouldMatch[i], inRange, testKey)
			}
		})
	}
}

func TestNewIndexDataStoreKey(t *testing.T) {
	fields := []IndexedField{
		{Value: client.NewNormalString("test"), Descending: false},
		{Value: client.NewNormalInt(42), Descending: true},
	}

	key := NewIndexDataStoreKey(123, 456, fields)

	assert.Equal(t, uint32(123), key.CollectionShortID)
	assert.Equal(t, uint32(456), key.IndexID)
	assert.Equal(t, 2, len(key.Fields))
	assert.True(t, key.Fields[0].Value.Equal(client.NewNormalString("test")))
	assert.False(t, key.Fields[0].Descending)
	assert.True(t, key.Fields[1].Value.Equal(client.NewNormalInt(42)))
	assert.True(t, key.Fields[1].Descending)
}

func TestIndexDataStoreKey_Bytes(t *testing.T) {
	tests := []struct {
		name string
		key  IndexDataStoreKey
		want string
	}{
		{
			name: "empty key",
			key:  IndexDataStoreKey{},
			want: "",
		},
		{
			name: "collection only",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
			},
		},
		{
			name: "collection and index",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
			},
		},
		{
			name: "with fields",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes1 := tt.key.Bytes()
			bytes2 := tt.key.Bytes()

			assert.Equal(t, bytes1, bytes2, "Bytes() should return consistent results")

			if tt.name == "empty key" {
				assert.Empty(t, bytes1)
			} else {
				assert.NotEmpty(t, bytes1)
			}
		})
	}
}

func TestIndexDataStoreKey_ToDS(t *testing.T) {
	key := IndexDataStoreKey{
		CollectionShortID: 123,
		IndexID:           456,
		Fields: []IndexedField{
			{Value: client.NewNormalString("test"), Descending: false},
		},
	}

	dsKey := key.ToDS()
	assert.NotNil(t, dsKey)

	assert.Equal(t, key.ToString(), dsKey.String())
}

func TestIndexDataStoreKey_ToString(t *testing.T) {
	tests := []struct {
		name string
		key  IndexDataStoreKey
	}{
		{
			name: "empty key",
			key:  IndexDataStoreKey{},
		},
		{
			name: "with collection",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
			},
		},
		{
			name: "with collection and index",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
			},
		},
		{
			name: "with fields",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
					{Value: client.NewNormalInt(42), Descending: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.key.ToString()

			assert.Equal(t, string(tt.key.Bytes()), str)

			if tt.key.CollectionShortID != 0 {
				assert.True(t, len(str) > 0 && str[0] == '/')
			}
		})
	}
}

func TestIndexDataStoreKey_Equal(t *testing.T) {
	tests := []struct {
		name  string
		key1  IndexDataStoreKey
		key2  IndexDataStoreKey
		equal bool
	}{
		{
			name:  "empty keys",
			key1:  IndexDataStoreKey{},
			key2:  IndexDataStoreKey{},
			equal: true,
		},
		{
			name: "same keys",
			key1: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
			key2: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
			equal: true,
		},
		{
			name:  "different collection",
			key1:  IndexDataStoreKey{CollectionShortID: 123},
			key2:  IndexDataStoreKey{CollectionShortID: 124},
			equal: false,
		},
		{
			name:  "different index",
			key1:  IndexDataStoreKey{CollectionShortID: 123, IndexID: 456},
			key2:  IndexDataStoreKey{CollectionShortID: 123, IndexID: 457},
			equal: false,
		},
		{
			name: "different field count",
			key1: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
			key2: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
					{Value: client.NewNormalInt(42), Descending: false},
				},
			},
			equal: false,
		},
		{
			name: "different field value",
			key1: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test1"), Descending: false},
				},
			},
			key2: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test2"), Descending: false},
				},
			},
			equal: false,
		},
		{
			name: "different field descending",
			key1: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
			key2: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: true},
				},
			},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key1.Equal(tt.key2)
			assert.Equal(t, tt.equal, result)

			// Equal should be symmetric
			assert.Equal(t, tt.equal, tt.key2.Equal(tt.key1))
		})
	}
}

func TestIndexDataStoreKey_EncodeDecode(t *testing.T) {
	indexDesc := &client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Descending: false},
			{Descending: true},
		},
	}

	fieldDefs := []client.FieldDefinition{
		{Kind: client.FieldKind_NILLABLE_STRING},
		{Kind: client.FieldKind_NILLABLE_INT},
	}

	tests := []struct {
		name      string
		key       IndexDataStoreKey
		wantError bool
	}{
		{
			name: "empty key encoding",
			key:  IndexDataStoreKey{},
		},
		{
			name: "collection only",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
			},
		},
		{
			name: "collection and index",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
			},
		},
		{
			name: "with single field",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
				},
			},
		},
		{
			name: "with multiple fields",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
					{Value: client.NewNormalInt(42), Descending: true},
				},
			},
		},
		{
			name: "with nil value",
			key: IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: func() client.NormalValue {
						v, _ := client.NewNormalNil(client.FieldKind_NILLABLE_STRING)
						return v
					}(), Descending: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeIndexDataStoreKey(&tt.key)

			if tt.key.CollectionShortID == 0 {
				assert.Empty(t, encoded)
				return
			}

			decoded, err := DecodeIndexDataStoreKey(encoded, indexDesc, fieldDefs)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.key.CollectionShortID, decoded.CollectionShortID)
			assert.Equal(t, tt.key.IndexID, decoded.IndexID)
			assert.Equal(t, len(tt.key.Fields), len(decoded.Fields))

			for i := range tt.key.Fields {
				assert.True(t, tt.key.Fields[i].Value.Equal(decoded.Fields[i].Value),
					"Field %d value mismatch", i)
				assert.Equal(t, tt.key.Fields[i].Descending, decoded.Fields[i].Descending,
					"Field %d descending mismatch", i)
			}
		})
	}
}

func TestIndexDataStoreKey_Decode(t *testing.T) {
	indexDesc := &client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Descending: false},
		},
	}

	fieldDefs := []client.FieldDefinition{
		{Kind: client.FieldKind_NILLABLE_STRING},
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: ErrEmptyKey,
		},
		{
			name:    "invalid start",
			data:    []byte("invalid"),
			wantErr: ErrInvalidKey,
		},
		{
			name:    "missing separator",
			data:    []byte("/123"),
			wantErr: ErrInvalidKey,
		},
		{
			name: "valid with doc ID field",
			data: EncodeIndexDataStoreKey(&IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
					{Value: client.NewNormalString("docID"), Descending: false},
				},
			}),
			wantErr: nil, // This is valid - second field is treated as doc ID
		},
		{
			name: "too many fields",
			data: EncodeIndexDataStoreKey(&IndexDataStoreKey{
				CollectionShortID: 123,
				IndexID:           456,
				Fields: []IndexedField{
					{Value: client.NewNormalString("test"), Descending: false},
					{Value: client.NewNormalString("docID"), Descending: false},
					{Value: client.NewNormalString("extra"), Descending: false},
				},
			}),
			wantErr: ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeIndexDataStoreKey(tt.data, indexDesc, fieldDefs)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
