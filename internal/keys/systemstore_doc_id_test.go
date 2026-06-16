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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

func TestSystemstoreDocIDKeys(t *testing.T) {
	const (
		collectionShortID uint32 = 42
		shortDocID        uint64 = 7
		publicDocID              = "bae-public-doc"
		fieldCID                 = "bafy-field-cid"
	)
	collectionSegment := string(encoding.EncodeUvarintAscending(nil, uint64(collectionShortID)))
	shortDocIDSegment := string(EncodeDocShortID(shortDocID))

	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "short to public",
			key:  NewShortIDToDocIDKey(collectionShortID, shortDocID),
			want: "/docid/" + collectionSegment + "/s/" + shortDocIDSegment,
		},
		{
			name: "public to short",
			key:  NewDocIDToShortIDKey(collectionShortID, publicDocID),
			want: "/docid/" + collectionSegment + "/p/" + publicDocID,
		},
		{
			name: "node public to short",
			key:  NewNodeDocIDToShortIDKey(publicDocID),
			want: "/docid/n/p/" + publicDocID,
		},
		{
			name: "node short to public",
			key:  NewNodeShortIDToDocIDKey(shortDocID),
			want: "/docid/n/s/" + shortDocIDSegment,
		},
		{
			name: "block to public",
			key:  NewBlockCIDToDocIDKey(collectionShortID, fieldCID, publicDocID),
			want: "/docid/" + collectionSegment + "/b/" + fieldCID + "/" + publicDocID,
		},
		{
			name: "public to block",
			key:  NewDocIDToBlockCIDKey(collectionShortID, publicDocID, fieldCID),
			want: "/docid/" + collectionSegment + "/pb/" + publicDocID + "/" + fieldCID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, test.key.ToString())
			require.Equal(t, test.want, string(test.key.Bytes()))
			require.Equal(t, test.want, test.key.ToDS().String())
		})
	}
}
