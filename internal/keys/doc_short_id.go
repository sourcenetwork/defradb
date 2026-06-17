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

func encodeShortID(shortID uint32) []byte {
	if shortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, uint64(shortID))
}

func decodeShortID(data []byte) (uint32, error) {
	rest, shortID, err := decodeShortIDPrefix(data)
	if err != nil {
		return 0, err
	}
	if len(rest) > 0 {
		return 0, ErrInvalidKey
	}
	return shortID, nil
}

func decodeShortIDPrefix(data []byte) ([]byte, uint32, error) {
	if len(data) == 0 {
		return nil, 0, ErrInvalidKey
	}
	rest, shortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return nil, 0, err
	}
	if shortID == 0 || shortID > math.MaxUint32 {
		return nil, 0, ErrInvalidKey
	}
	return rest, uint32(shortID), nil
}

// EncodeCollectionShortID returns the sortable path encoding for a local collection ID.
func EncodeCollectionShortID(collectionShortID uint32) []byte {
	return encodeShortID(collectionShortID)
}

// DecodeCollectionShortID decodes a local collection ID from a single encoded key segment.
func DecodeCollectionShortID(data []byte) (uint32, error) {
	return decodeShortID(data)
}

// DecodeCollectionShortIDPrefix decodes a local collection ID from the start of data.
func DecodeCollectionShortIDPrefix(data []byte) ([]byte, uint32, error) {
	return decodeShortIDPrefix(data)
}

// EncodeDocShortID returns the sortable path encoding for a local document storage ID.
func EncodeDocShortID(docShortID uint32) []byte {
	return encodeShortID(docShortID)
}

// DecodeDocShortID decodes a local document storage ID from a single encoded key segment.
func DecodeDocShortID(data []byte) (uint32, error) {
	return decodeShortID(data)
}

// DecodeDocShortIDPrefix decodes a local document storage ID from the start of data.
func DecodeDocShortIDPrefix(data []byte) ([]byte, uint32, error) {
	return decodeShortIDPrefix(data)
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
