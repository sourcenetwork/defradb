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
	"math"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

const slashEncodedShortID = 303

func TestShortIDPathParsers_HandleSlashInEncodedID(t *testing.T) {
	require.Contains(t, EncodeDocShortID(slashEncodedShortID), byte('/'))

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
		DocShortID: slashEncodedShortID,
		FieldID:    "2",
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
		0,
		[]IndexedField{
			{Value: client.NewNormalInt(33)},
		},
	)
	indexKey.DocShortID = slashEncodedShortID
	decodedIndexKey, err := DecodeIndexDataStoreKey(indexKey.Bytes(), indexDesc, fieldDefs, 0)
	require.NoError(t, err)
	require.True(t, indexKey.Equal(decodedIndexKey))
}

func TestDecodeDocShortIDPrefix(t *testing.T) {
	encoded := EncodeDocShortID(slashEncodedShortID)
	data := append(append([]byte{}, encoded...), '/', '2')

	rest, docShortID, err := DecodeDocShortIDPrefix(data)

	require.NoError(t, err)
	require.Equal(t, uint64(slashEncodedShortID), docShortID)
	require.True(t, bytes.Equal([]byte{'/', '2'}, rest))
}

func TestDocShortID_RoundTrip(t *testing.T) {
	for _, docShortID := range []uint64{1, 109, 110, slashEncodedShortID, 16384, math.MaxUint32, math.MaxUint64} {
		decoded, err := DecodeDocShortID(EncodeDocShortID(docShortID))
		require.NoError(t, err)
		require.Equal(t, docShortID, decoded)
	}
}

func TestDocShortID_ZeroIsReserved(t *testing.T) {
	require.Nil(t, EncodeDocShortID(0))

	_, err := DecodeDocShortID(nil)
	require.ErrorIs(t, err, ErrInvalidKey)

	_, err = DecodeDocShortID(encoding.EncodeUvarintAscending(nil, 0))
	require.ErrorIs(t, err, ErrInvalidKey)
}

func TestDocShortID_RejectsTrailingBytes(t *testing.T) {
	data := append(EncodeDocShortID(42), 'x')
	_, err := DecodeDocShortID(data)
	require.ErrorIs(t, err, ErrInvalidKey)
}

func TestDocRef_RoundTrip(t *testing.T) {
	ref := DocRef{CollectionShortID: slashEncodedShortID, DocShortID: math.MaxUint64}
	decoded, err := DecodeDocRef(EncodeDocRef(ref.CollectionShortID, ref.DocShortID))
	require.NoError(t, err)
	require.Equal(t, ref, decoded)
}

func TestDocRef_ZeroComponentsEncodeToNil(t *testing.T) {
	require.Nil(t, EncodeDocRef(0, 7))
	require.Nil(t, EncodeDocRef(7, 0))
}

func TestDocRef_DecodeRejectsInvalid(t *testing.T) {
	zeroCollection := append(encoding.EncodeUvarintAscending(nil, 0), EncodeDocShortID(7)...)
	_, err := DecodeDocRef(zeroCollection)
	require.ErrorIs(t, err, ErrInvalidKey)

	overflowCollection := append(encoding.EncodeUvarintAscending(nil, math.MaxUint32+1), EncodeDocShortID(7)...)
	_, err = DecodeDocRef(overflowCollection)
	require.ErrorIs(t, err, ErrInvalidKey)
}

func TestDecodeCollectionShortIDPrefix_RejectsOverflow(t *testing.T) {
	_, _, err := DecodeCollectionShortIDPrefix(encoding.EncodeUvarintAscending(nil, math.MaxUint32+1))
	require.ErrorIs(t, err, ErrInvalidKey)
}
