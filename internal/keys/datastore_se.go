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
	"encoding/hex"
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

const (
	SE_PREFIX = "/se"
)

// DatastoreSE provides key generation for SE artifacts
type DatastoreSE struct {
	CollectionShortID uint32
	IndexID           string
	SearchTag         []byte
	DocID             string
}

var _ Key = (*DatastoreSE)(nil)
var _ Walkable = (*DatastoreSE)(nil)
var _ CollectionedKey = DatastoreSE{}

func (k DatastoreSE) Bytes() []byte {
	return []byte(k.ToString())
}

func (k DatastoreSE) ToString() string {
	var sb strings.Builder
	sb.WriteString(SE_PREFIX)

	if k.CollectionShortID != 0 {
		sb.WriteString("/")

		colIDBytes := encoding.EncodeUvarintAscending([]byte{}, uint64(k.CollectionShortID))
		sb.Write(colIDBytes)

		if k.IndexID != "" {
			sb.WriteString("/")
			sb.WriteString(k.IndexID)

			if len(k.SearchTag) > 0 {
				sb.WriteString("/")
				sb.WriteString(hex.EncodeToString(k.SearchTag))

				if k.DocID != "" {
					sb.WriteString("/")
					sb.WriteString(k.DocID)
				}
			}
		}
	}

	return sb.String()
}

func (k DatastoreSE) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k DatastoreSE) PrefixEnd() Walkable {
	return rawKey(bytesPrefixEnd(k.Bytes()))
}

// NewDatastoreSEFromString creates a DatastoreSE from a key string
func NewDatastoreSEFromString(key string) (DatastoreSE, error) {
	if key != SE_PREFIX && !strings.HasPrefix(key, SE_PREFIX+"/") {
		return DatastoreSE{}, errors.New("invalid SE key format")
	}

	k := DatastoreSE{}
	data := []byte(key[len(SE_PREFIX):])
	if len(data) == 0 {
		return k, nil
	}
	if data[0] != '/' {
		return DatastoreSE{}, errors.New("invalid SE key format")
	}
	data = data[1:]

	if len(data) > 0 && data[0] != '/' {
		rest, collectionShortID, err := DecodeCollectionShortIDPrefix(data)
		if err != nil {
			return DatastoreSE{}, err
		}
		k.CollectionShortID = collectionShortID
		data = rest
	}

	if len(data) == 0 {
		return k, nil
	}
	if data[0] != '/' {
		return DatastoreSE{}, errors.New("invalid SE key format")
	}

	parts := strings.Split(string(data[1:]), "/")
	if len(parts) > 0 {
		k.IndexID = parts[0]
	}

	if len(parts) > 1 {
		searchTag, err := hex.DecodeString(parts[1])
		if err != nil {
			return DatastoreSE{}, errors.Wrap("failed to decode search tag", err)
		}
		k.SearchTag = searchTag
	}

	if len(parts) > 2 {
		k.DocID = parts[2]
	}

	return k, nil
}

func (k DatastoreSE) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}
