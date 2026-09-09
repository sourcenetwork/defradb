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
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

// IndexedField contains information necessary for storing a single
// value of a field in an index.
type IndexedField struct {
	// Value is the value of the field in the index
	Value client.NormalValue
	// Descending is true if the field is sorted in descending order
	Descending bool
}

// IndexDataStoreKey is key of an indexed document in the database.
type IndexDataStoreKey struct {
	// CollectionShortID is the id of the collection
	CollectionShortID uint32
	// IndexID is the id of the index
	IndexID uint32
	// Epoch namespaces the entries of a single index build; a rebuild fills a fresh epoch
	// disjoint from the current one. Stored entries always carry an epoch of 1 or greater.
	//
	// Epoch 0 carries no component, forming a prefix over every epoch — used to scan or drop
	// the whole index.
	Epoch uint32
	// Fields is the values of the fields in the index
	Fields []IndexedField
	// DocShortID is the trailing suffix that makes equal index values unique.
	// It also lets index scans resolve the matching document without storing the full DocID.
	DocShortID uint64
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
func NewIndexDataStoreKey(collectionShortID, indexID, epoch uint32, fields []IndexedField) IndexDataStoreKey {
	return IndexDataStoreKey{
		CollectionShortID: collectionShortID,
		IndexID:           indexID,
		Epoch:             epoch,
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
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)(/[DocShortID])
// If while composing the string from left to right, a component
// is empty, the string is returned up to that point
func (k *IndexDataStoreKey) ToString() string {
	return string(k.Bytes())
}

// Equal returns true if the two keys are equal
func (k *IndexDataStoreKey) Equal(other IndexDataStoreKey) bool {
	if k.CollectionShortID != other.CollectionShortID || k.IndexID != other.IndexID ||
		k.Epoch != other.Epoch {
		return false
	}

	if len(k.Fields) != len(other.Fields) {
		return false
	}

	for i, field := range k.Fields {
		if !field.Value.Equal(other.Fields[i].Value) ||
			field.Descending != other.Fields[i].Descending {
			return false
		}
	}

	return k.DocShortID == other.DocShortID
}

// DecodeIndexDataStoreKey decodes a IndexDataStoreKey from bytes.
// It expects the input bytes is in the following format:
//
// /[CollectionID]/[IndexID](/[Epoch])(/[FieldValue]...)(/[DocShortID])
//
// Where [CollectionID], [IndexID] and [Epoch] are integers.
//
// The epoch component is indistinguishable from a field value by structure, so the caller
// supplies the epoch of the keyspace it is scanning; the decoder consumes a component only when
// it is non-zero.
//
// Field values are decoded to standardized Defra Go types per the field descriptions.
func DecodeIndexDataStoreKey(
	data []byte,
	indexDesc *client.IndexDescription,
	fields []client.CollectionFieldDescription,
	epoch uint32,
) (IndexDataStoreKey, error) {
	if len(data) == 0 {
		return IndexDataStoreKey{}, ErrEmptyKey
	}

	if data[0] != '/' {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, collectionShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return IndexDataStoreKey{}, err
	}

	key := IndexDataStoreKey{CollectionShortID: uint32(collectionShortID)}

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

	// The epoch component looks identical to a field value, so the decoder can't detect it on its
	// own. The caller passes the epoch of the keyspace it scanned: non-zero for real entries (read
	// it), zero for a whole-index prefix scan (skip it, the rest are field values).
	if epoch != 0 {
		if data[0] != '/' {
			return IndexDataStoreKey{}, ErrInvalidKey
		}
		data = data[1:]

		var ep uint64
		data, ep, err = encoding.DecodeUvarintAscending(data)
		if err != nil {
			return IndexDataStoreKey{}, err
		}
		key.Epoch = uint32(ep)

		if len(data) == 0 {
			return key, nil
		}
	}

	for len(data) > 0 {
		if data[0] != '/' {
			return IndexDataStoreKey{}, ErrInvalidKey
		}
		data = data[1:]

		i := len(key.Fields)
		descending := false
		if i < len(indexDesc.GetFields()) {
			descending = indexDesc.GetFields()[i].Descending
		} else if i > len(indexDesc.GetFields()) {
			return IndexDataStoreKey{}, ErrInvalidKey
		} else {
			if key.DocShortID != 0 {
				return IndexDataStoreKey{}, ErrInvalidKey
			}
			data, key.DocShortID, err = DecodeDocShortIDPrefix(data)
			if err != nil {
				return IndexDataStoreKey{}, err
			}
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

		// Epoch 0 carries no component, forming a prefix over every epoch of the index.
		if key.Epoch != 0 {
			b = append(b, '/')
			b = encoding.EncodeUvarintAscending(b, uint64(key.Epoch))
		}

		for _, field := range key.Fields {
			b = append(b, '/')
			b = encoding.EncodeFieldValue(b, field.Value, field.Descending)
		}
		if key.DocShortID != 0 {
			b = append(b, '/')
			b = append(b, EncodeDocShortID(key.DocShortID)...)
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
		Epoch:             k.Epoch,
		Fields:            newFields,
		DocShortID:        k.DocShortID,
		Offset:            k.Offset + 1,
	}
}

func (k *IndexDataStoreKey) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}
