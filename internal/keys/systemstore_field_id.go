// Copyright 2024 Democratized Data Foundation
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
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

// FieldID indexes field short ids by the full id.
type FieldID struct {
	CollectionShortID uint32
	FieldID           string
}

var _ Key = (*CollectionVersionKey)(nil)

func NewFieldID(collectionShortID uint32, fieldID string) FieldID {
	return FieldID{
		CollectionShortID: collectionShortID,
		FieldID:           fieldID,
	}
}

func NewFieldIDPrefix(collectionShortID uint32) FieldID {
	return FieldID{
		CollectionShortID: collectionShortID,
	}
}

func NewFieldIDFromBytes(key []byte) (FieldID, error) {
	if !bytes.HasPrefix(key, []byte(FIELD_SHORT_ID)) {
		return FieldID{}, ErrInvalidKey
	}

	key = bytes.TrimPrefix(key, []byte(FIELD_SHORT_ID))
	if len(key) == 0 {
		return FieldID{}, nil
	}
	if key[0] != '/' {
		return FieldID{}, ErrInvalidKey
	}
	key = key[1:]

	key, colID, err := encoding.DecodeUvarintAscending(key)
	if err != nil {
		return FieldID{}, err
	}

	var fieldID string
	if len(key) > 1 {
		if key[0] == '/' {
			key = key[1:]
		}
		fieldID = strings.TrimSuffix(string(key), "/")
	}

	return FieldID{
		CollectionShortID: uint32(colID),
		FieldID:           fieldID,
	}, nil
}

func (k FieldID) ToString() string {
	result := FIELD_SHORT_ID

	if k.CollectionShortID != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionShortID))
	}

	if k.FieldID != "" {
		result = result + "/" + k.FieldID
	}

	return result
}

func (k FieldID) Bytes() []byte {
	result := []byte(FIELD_SHORT_ID)

	if k.CollectionShortID != 0 {
		result = append(result, encoding.EncodeUvarintAscending([]byte{'/'}, uint64(k.CollectionShortID))...)
	}

	if k.FieldID != "" {
		result = append(result, '/')
		result = append(result, []byte(k.FieldID)...)
	}

	return result
}

func (k FieldID) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
