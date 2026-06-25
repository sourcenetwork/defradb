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

// DocRef is the systemstore value used to resolve a DocID to local storage.
// CollectionShortID is stored with the node-unique DocShortID so callers can
// recover the document's collection from only its DocID.
type DocRef struct {
	CollectionShortID uint32
	DocShortID        uint64
}

// EncodeDocRef encodes a local collection/doc pair as a compact systemstore value.
func EncodeDocRef(collectionShortID uint32, docShortID uint64) []byte {
	if collectionShortID == 0 || docShortID == 0 {
		return nil
	}
	result := encodeShortID(uint64(collectionShortID))
	return append(result, EncodeDocShortID(docShortID)...)
}

// DecodeDocRef decodes a local collection/doc pair from a compact systemstore value.
func DecodeDocRef(data []byte) (DocRef, error) {
	rest, collectionShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return DocRef{}, err
	}
	docShortID, err := DecodeDocShortID(rest)
	if err != nil {
		return DocRef{}, err
	}
	if collectionShortID == 0 || collectionShortID > math.MaxUint32 {
		return DocRef{}, ErrInvalidKey
	}
	return DocRef{
		CollectionShortID: uint32(collectionShortID),
		DocShortID:        docShortID,
	}, nil
}
