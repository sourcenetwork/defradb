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
	"math"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

// LocalDocID is this node's storage address for a public document ID.
type LocalDocID struct {
	CollectionShortID uint32
	DocShortID        uint32
}

// EncodeCollectionShortID returns the sortable path encoding for a local collection ID.
func EncodeCollectionShortID(collectionShortID uint32) []byte {
	if collectionShortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, uint64(collectionShortID))
}

// DecodeCollectionShortID decodes a local collection ID from a single encoded key segment.
func DecodeCollectionShortID(data []byte) (uint32, error) {
	rest, collectionShortID, err := DecodeCollectionShortIDPrefix(data)
	if err != nil {
		return 0, err
	}
	if len(rest) > 0 {
		return 0, ErrInvalidKey
	}
	return collectionShortID, nil
}

// DecodeCollectionShortIDPrefix decodes a local collection ID from the start of data.
func DecodeCollectionShortIDPrefix(data []byte) ([]byte, uint32, error) {
	if len(data) == 0 {
		return nil, 0, ErrInvalidKey
	}
	rest, collectionShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return nil, 0, err
	}
	if collectionShortID == 0 || collectionShortID > math.MaxUint32 {
		return nil, 0, ErrInvalidKey
	}
	return rest, uint32(collectionShortID), nil
}

// EncodeDocShortID returns the sortable path encoding for a local document storage ID.
func EncodeDocShortID(docShortID uint32) []byte {
	if docShortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, uint64(docShortID))
}

// DecodeDocShortID decodes a local document storage ID from a single encoded key segment.
func DecodeDocShortID(data []byte) (uint32, error) {
	rest, docShortID, err := DecodeDocShortIDPrefix(data)
	if err != nil {
		return 0, err
	}
	if len(rest) > 0 {
		return 0, ErrInvalidKey
	}
	return docShortID, nil
}

// DecodeDocShortIDPrefix decodes a local document storage ID from the start of data.
func DecodeDocShortIDPrefix(data []byte) ([]byte, uint32, error) {
	if len(data) == 0 {
		return nil, 0, ErrInvalidKey
	}
	rest, docShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return nil, 0, err
	}
	if docShortID == 0 || docShortID > math.MaxUint32 {
		return nil, 0, ErrInvalidKey
	}
	return rest, uint32(docShortID), nil
}

// EncodeLocalDocID encodes a local collection/doc pair as a compact systemstore value.
func EncodeLocalDocID(collectionShortID uint32, docShortID uint32) []byte {
	if collectionShortID == 0 || docShortID == 0 {
		return nil
	}
	result := EncodeCollectionShortID(collectionShortID)
	return append(result, EncodeDocShortID(docShortID)...)
}

// DecodeLocalDocID decodes a local collection/doc pair from a compact systemstore value.
func DecodeLocalDocID(data []byte) (LocalDocID, error) {
	rest, collectionShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return LocalDocID{}, err
	}
	docShortID, err := DecodeDocShortID(rest)
	if err != nil {
		return LocalDocID{}, err
	}
	if collectionShortID == 0 || collectionShortID > math.MaxUint32 {
		return LocalDocID{}, ErrInvalidKey
	}
	return LocalDocID{
		CollectionShortID: uint32(collectionShortID),
		DocShortID:        docShortID,
	}, nil
}
