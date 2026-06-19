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
		docShortID        uint32 = 7
		docID                    = "bae-public-doc"
		fieldCID                 = "bafy-field-cid"
	)
	collectionSegment := string(encoding.EncodeUvarintAscending(nil, uint64(collectionShortID)))
	docShortIDSegment := string(EncodeDocShortID(docShortID))

	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "short to public",
			key:  NewShortIDToDocIDKey(collectionShortID, docShortID),
			want: "/docid/" + collectionSegment + "/s/" + docShortIDSegment,
		},
		{
			name: "node public to short",
			key:  NewNodeDocIDToShortIDKey(docID),
			want: "/docid/n/p/" + docID,
		},
		{
			name: "block to public",
			key:  NewBlockCIDToDocIDKey(collectionShortID, fieldCID),
			want: "/docid/b/" + fieldCID + "/" + collectionSegment,
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
