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
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	PEERSTORE_SE_RETRY_PREFIX = "/se-retry"
)

// PeerstoreSERetry provides key generation for SE retry information
type PeerstoreSERetry struct {
	PeerID       string
	CollectionID string
	DocID        string
}

var _ Key = (*PeerstoreSERetry)(nil)

func NewPeerstoreSERetry(peerID, collectionID, docID string) PeerstoreSERetry {
	return PeerstoreSERetry{
		PeerID:       peerID,
		CollectionID: collectionID,
		DocID:        docID,
	}
}

func (k PeerstoreSERetry) ToString() string {
	var sb strings.Builder
	sb.WriteString(PEERSTORE_SE_RETRY_PREFIX)
	
	if k.PeerID != "" {
		sb.WriteString("/")
		sb.WriteString(k.PeerID)
		
		if k.CollectionID != "" {
			sb.WriteString("/")
			sb.WriteString(k.CollectionID)
			
			if k.DocID != "" {
				sb.WriteString("/")
				sb.WriteString(k.DocID)
			}
		}
	}
	
	return sb.String()
}

func (k PeerstoreSERetry) Bytes() []byte {
	return []byte(k.ToString())
}

func (k PeerstoreSERetry) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func NewPeerstoreSERetryFromString(key string) (PeerstoreSERetry, error) {
	parts := strings.Split(key, "/")
	// Expected format: /se-retry/<peerID>/<collectionID>/<docID>
	if len(parts) < 5 || parts[1] != "se-retry" {
		return PeerstoreSERetry{}, errors.New("invalid SE retry key format")
	}

	return PeerstoreSERetry{
		PeerID:       parts[2],
		CollectionID: parts[3],
		DocID:        parts[4],
	}, nil
}
