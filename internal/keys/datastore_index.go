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
	"math"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

// DocIDIndexID is reserved for the internal public DocID to short doc ID index.
const DocIDIndexID = math.MaxUint32

// IndexedField contains information necessary for storing a single
// value of a field in an index.
type IndexedField struct {
	// Value is the value of the field in the index
	Value client.NormalValue
	// Descending is true if the field is sorted in descending order
	Descending bool
	docShortID uint32
}

// IndexDataStoreKey is key of an indexed document in the database.
type IndexDataStoreKey struct {
	// CollectionShortID is the id of the collection
	CollectionShortID uint32
	// IndexID is the id of the index
	IndexID uint32
	// Fields is the values of the fields in the index
	Fields []IndexedField
	// Offset can be set in order to control how many times `bytesPrefixEnd` is called when this `IndexDataStoreKey`
	// is serialized.
	//
	// This allows `bytesPrefixEnd` to be managed before serialization, allowing the `bytesPrefixEnd`'ed key to be
	// passed into strongly typed interfaces, such as `Keyedstore`.
	Offset uint64
}

var _ Walkable = (*IndexDataStoreKey)(nil)
var _ CollectionedKey = (*IndexDataStoreKey)(nil)

// NewIndexDataStoreKey creates a new IndexDataStoreKey from a collection ID, index ID and fields.
// It also validates values of the fields.
func NewIndexDataStoreKey(collectionShortID, indexID uint32, fields []IndexedField) IndexDataStoreKey {
	return IndexDataStoreKey{
		CollectionShortID: collectionShortID,
		IndexID:           indexID,
		Fields:            fields,
	}
}

// Bytes returns the byte representation of the key
func (k *IndexDataStoreKey) Bytes() []byte {
	return EncodeIndexDataStoreKey(k)
}

// ToDS returns the datastore key
func (k *IndexDataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// ToString returns the string representation of the key
// It is in the following format:
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)
// If while composing the string from left to right, a component
// is empty, the string is returned up to that point
func (k *IndexDataStoreKey) ToString() string {
	return string(k.Bytes())
}

// Equal returns true if the two keys are equal
func (k *IndexDataStoreKey) Equal(other IndexDataStoreKey) bool {
	if k.CollectionShortID != other.CollectionShortID || k.IndexID != other.IndexID {
		return false
	}

	if len(k.Fields) != len(other.Fields) {
		return false
	}

	for i, field := range k.Fields {
		if !field.Value.Equal(other.Fields[i].Value) ||
			field.Descending != other.Fields[i].Descending ||
			field.docShortID != other.Fields[i].docShortID {
			return false
		}
	}

	return true
}

// NewDocShortIDIndexedField returns the hidden document suffix used by non-unique index keys.
func NewDocShortIDIndexedField(shortDocID uint32) IndexedField {
	return IndexedField{
		Value:      client.NewNormalBytes(EncodeDocShortID(shortDocID)),
		docShortID: shortDocID,
	}
}

// DecodeIndexDataStoreKey decodes a IndexDataStoreKey from bytes.
// It expects the input bytes is in the following format:
//
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)
//
// Where [CollectionID] and [IndexID] are integers
//
// All values of the fields are converted to standardized Defra Go type
// according to fields description.
func DecodeIndexDataStoreKey(
	data []byte,
	indexDesc *client.IndexDescription,
	fields []client.CollectionFieldDescription,
) (IndexDataStoreKey, error) {
	if len(data) == 0 {
		return IndexDataStoreKey{}, ErrEmptyKey
	}

	if data[0] != '/' {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, colID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return IndexDataStoreKey{}, err
	}

	key := IndexDataStoreKey{CollectionShortID: uint32(colID)}

	if len(data) == 0 {
		return key, nil
	}

	if data[0] != '/' {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, indID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return IndexDataStoreKey{}, err
	}
	key.IndexID = uint32(indID)

	if len(data) == 0 {
		return key, nil
	}

	for len(data) > 0 {
		if data[0] != '/' {
			return IndexDataStoreKey{}, ErrInvalidKey
		}
		data = data[1:]

		i := len(key.Fields)
		descending := false
		if i < len(indexDesc.Fields) {
			descending = indexDesc.Fields[i].Descending
		} else if i > len(indexDesc.Fields) {
			return IndexDataStoreKey{}, ErrInvalidKey
		} else {
			shortDocIDEnd := len(data)
			if slash := bytes.IndexByte(data, '/'); slash >= 0 {
				shortDocIDEnd = slash
			}
			shortDocID, err := DecodeDocShortID(data[:shortDocIDEnd])
			if err != nil {
				return IndexDataStoreKey{}, err
			}
			key.Fields = append(key.Fields, NewDocShortIDIndexedField(shortDocID))
			data = data[shortDocIDEnd:]
			continue
		}

		kind := fields[i].Kind
		if kind != nil && kind.IsArray() {
			if arrKind, ok := kind.(client.ScalarArrayKind); ok {
				kind = arrKind.SubKind()
			}
		}

		var val client.NormalValue
		data, val, err = encoding.DecodeFieldValue(data, descending, kind)
		if err != nil {
			return IndexDataStoreKey{}, err
		}

		key.Fields = append(key.Fields, IndexedField{Value: val, Descending: descending})
	}

	return key, nil
}

// EncodeIndexDataStoreKey encodes a IndexDataStoreKey to bytes to be stored as a key
// for secondary indexes.
func EncodeIndexDataStoreKey(key *IndexDataStoreKey) []byte {
	if key.CollectionShortID == 0 {
		return []byte{}
	}

	b := encoding.EncodeUvarintAscending([]byte{'/'}, uint64(key.CollectionShortID))

	if key.IndexID != 0 {
		b = append(b, '/')
		b = encoding.EncodeUvarintAscending(b, uint64(key.IndexID))

		for _, field := range key.Fields {
			b = append(b, '/')
			if field.docShortID != 0 {
				b = append(b, EncodeDocShortID(field.docShortID)...)
			} else {
				b = encoding.EncodeFieldValue(b, field.Value, field.Descending)
			}
		}
	}

	for i := 0; i < int(key.Offset); i++ {
		b = bytesPrefixEnd(b)
	}

	return b
}

// PrefixEnd returns a key that would sort immediately after all keys with this prefix.
// It returns a key such that all keys with the prefix are >= k and < k.PrefixEnd().
// This is implemented by encoding the key to bytes and incrementing it.
func (k *IndexDataStoreKey) PrefixEnd() Walkable {
	newFields := make([]IndexedField, len(k.Fields))
	copy(newFields, k.Fields)

	return &IndexDataStoreKey{
		CollectionShortID: k.CollectionShortID,
		IndexID:           k.IndexID,
		Fields:            newFields,
		Offset:            k.Offset + 1,
	}
}

func (k *IndexDataStoreKey) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}
