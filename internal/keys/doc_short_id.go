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

// Short IDs are encoded as sortable path bytes for datastore/headstore keys.
// Keep the encoding here so every key type sorts the same way. Do not split
// encoded IDs on '/', because valid encoded IDs may contain that byte.

func encodeShortID(shortID uint64) []byte {
	if shortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, shortID)
}

func decodeShortID(data []byte) (uint64, error) {
	rest, shortID, err := decodeShortIDPrefix(data)
	if err != nil {
		return 0, err
	}
	if len(rest) > 0 {
		return 0, ErrInvalidKey
	}
	return shortID, nil
}

// decodeShortIDPrefix returns the unconsumed suffix after one encoded short ID.
func decodeShortIDPrefix(data []byte) ([]byte, uint64, error) {
	if len(data) == 0 {
		return nil, 0, ErrInvalidKey
	}
	rest, shortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return nil, 0, err
	}
	if shortID == 0 {
		return nil, 0, ErrInvalidKey
	}
	return rest, shortID, nil
}

// DecodeCollectionShortIDPrefix decodes a collection short ID from a key prefix.
func DecodeCollectionShortIDPrefix(data []byte) ([]byte, uint32, error) {
	rest, collectionShortID, err := decodeShortIDPrefix(data)
	if err != nil {
		return nil, 0, err
	}
	if collectionShortID > math.MaxUint32 {
		return nil, 0, ErrInvalidKey
	}
	return rest, uint32(collectionShortID), nil
}

// EncodeDocShortID returns the encoded path form of a document short ID.
func EncodeDocShortID(docShortID uint64) []byte {
	return encodeShortID(docShortID)
}

// DecodeDocShortID decodes a document short ID from one encoded value.
func DecodeDocShortID(data []byte) (uint64, error) {
	return decodeShortID(data)
}

// DecodeDocShortIDPrefix decodes a document short ID from a key prefix.
func DecodeDocShortIDPrefix(data []byte) ([]byte, uint64, error) {
	return decodeShortIDPrefix(data)
}
