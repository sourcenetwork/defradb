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
)

func TestSystemstoreDocIDKeys(t *testing.T) {
	const (
		collectionShortID uint32 = 42
		docShortID        uint64 = 7
		docID                    = "bae-doc"
		fieldCID                 = "bafy-field-cid"
	)
	docShortIDSegment := string(EncodeDocShortID(docShortID))

	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "short to doc id",
			key:  NewShortIDToDocIDKey(docShortID),
			want: "/d/s/" + docShortIDSegment,
		},
		{
			name: "doc id to local doc ref",
			key:  NewDocIDToDocRefKey(docID),
			want: "/d/p/" + docID,
		},
		{
			name: "doc ref to doc id",
			key:  NewDocRefToDocIDKey(docShortID, docID),
			want: "/d/r/" + docShortIDSegment + "/" + docID,
		},
		{
			name: "block to doc id",
			key:  NewBlockCIDToDocIDKey(fieldCID, docID),
			want: "/d/b/" + fieldCID + "/" + docID,
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
