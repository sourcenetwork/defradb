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
)

const (
	SE_PREFIX = "/se"
)

// DatastoreSE provides key generation for SE artifacts
type DatastoreSE struct {
	CollectionID string
	IndexID      string
	SearchTag    []byte
	DocID        string
}

var _ Key = (*DatastoreSE)(nil)

func (k DatastoreSE) Bytes() []byte {
	return []byte(k.ToString())
}

func (k DatastoreSE) ToString() string {
	var sb strings.Builder
	sb.WriteString(SE_PREFIX)

	if k.CollectionID != "" {
		sb.WriteString("/")
		sb.WriteString(k.CollectionID)

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

// NewDatastoreSEFromString creates a DatastoreSE from a key string
func NewDatastoreSEFromString(key string) (DatastoreSE, error) {
	parts := strings.Split(key, "/")
	// Expected format: /se/<collectionID>/<indexID>/<searchTag>/<docID>
	if len(parts) < 2 || parts[1] != "se" {
		return DatastoreSE{}, errors.New("invalid SE key format")
	}

	k := DatastoreSE{}
	
	if len(parts) > 2 {
		k.CollectionID = parts[2]
	}
	
	if len(parts) > 3 {
		k.IndexID = parts[3]
	}
	
	if len(parts) > 4 {
		searchTag, err := hex.DecodeString(parts[4])
		if err != nil {
			return DatastoreSE{}, errors.Wrap("failed to decode search tag", err)
		}
		k.SearchTag = searchTag
	}
	
	if len(parts) > 5 {
		k.DocID = parts[5]
	}
	
	return k, nil
}
