// Copyright 2026 Democratized Data Foundation
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
	"bytes"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

const slashEncodedShortID uint32 = 303

func TestShortIDPathParsers_HandleSlashInEncodedID(t *testing.T) {
	require.Contains(t, EncodeDocShortID(slashEncodedShortID), byte('/'))
	require.Contains(t, EncodeCollectionShortID(slashEncodedShortID), byte('/'))

	docKey := DataStoreKey{
		CollectionShortID: slashEncodedShortID,
		InstanceType:      ValueKey,
		DocShortID:        slashEncodedShortID,
		FieldID:           "2",
	}
	decodedDocKey, err := DecodeDataStoreKey(docKey.Bytes())
	require.NoError(t, err)
	require.Equal(t, docKey, decodedDocKey)

	primaryKey := PrimaryDataStoreKey{
		CollectionShortID: slashEncodedShortID,
		DocShortID:        slashEncodedShortID,
	}
	decodedPrimaryKey, err := NewPrimaryDataStoreKey(primaryKey.ToString())
	require.NoError(t, err)
	require.Equal(t, primaryKey, decodedPrimaryKey)

	headKey := HeadstoreDocKey{
		CollectionShortID: slashEncodedShortID,
		DocShortID:        slashEncodedShortID,
		FieldID:           "2",
		Cid: cid.MustParse(
			"bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
		),
	}
	decodedHeadKey, err := NewHeadstoreDocKey(headKey.ToString())
	require.NoError(t, err)
	require.Equal(t, headKey, decodedHeadKey)

	indexDesc := &client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{}},
	}
	fieldDefs := []client.CollectionFieldDescription{
		{Kind: client.FieldKind_NILLABLE_INT},
	}
	indexKey := NewIndexDataStoreKey(
		slashEncodedShortID,
		1,
		[]IndexedField{
			{Value: client.NewNormalInt(33)},
			NewDocShortIDIndexedField(slashEncodedShortID),
		},
	)
	decodedIndexKey, err := DecodeIndexDataStoreKey(indexKey.Bytes(), indexDesc, fieldDefs)
	require.NoError(t, err)
	require.True(t, indexKey.Equal(decodedIndexKey))
}

func TestDecodeDocShortIDPrefix(t *testing.T) {
	encoded := EncodeDocShortID(slashEncodedShortID)
	data := append(append([]byte{}, encoded...), '/', '2')

	rest, shortDocID, err := DecodeDocShortIDPrefix(data)

	require.NoError(t, err)
	require.Equal(t, slashEncodedShortID, shortDocID)
	require.True(t, bytes.Equal([]byte{'/', '2'}, rest))
}
