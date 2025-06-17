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
