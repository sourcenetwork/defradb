// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestSetDocumentIDs(t *testing.T) {
	docIDs := []client.DocID{
		client.NewDocIDV0(blocks.NewBlock([]byte("doc one")).Cid()),
		client.NewDocIDV0(blocks.NewBlock([]byte("doc two")).Cid()),
	}

	docs := []*client.Document{{}, {}}
	require.NoError(t, setDocumentIDs(docs, []string{docIDs[0].String(), docIDs[1].String()}))

	require.Equal(t, docIDs[0].String(), docs[0].ID().String())
	require.Equal(t, docIDs[1].String(), docs[1].ID().String())
}

func TestSetDocumentIDsReturnsError(t *testing.T) {
	docID := client.NewDocIDV0(blocks.NewBlock([]byte("doc")).Cid())

	tests := []struct {
		name   string
		docs   []*client.Document
		docIDs []string
	}{
		{
			name:   "too few ids",
			docs:   []*client.Document{{}, {}},
			docIDs: []string{docID.String()},
		},
		{
			name:   "too many ids",
			docs:   []*client.Document{{}},
			docIDs: []string{docID.String(), docID.String()},
		},
		{
			name:   "invalid doc id",
			docs:   []*client.Document{{}},
			docIDs: []string{"not-a-doc-id"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, setDocumentIDs(test.docs, test.docIDs))
		})
	}
}
