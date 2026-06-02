// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import (
	"encoding/json"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestSetDocumentIDsFromJSON(t *testing.T) {
	docIDs := []client.DocID{
		client.NewDocIDV0(blocks.NewBlock([]byte("doc one")).Cid()),
		client.NewDocIDV0(blocks.NewBlock([]byte("doc two")).Cid()),
	}
	rawDocIDs, err := json.Marshal([]string{docIDs[0].String(), docIDs[1].String()})
	require.NoError(t, err)

	docs := []*client.Document{{}, {}}
	require.NoError(t, setDocumentIDsFromJSON(docs, rawDocIDs))

	require.Equal(t, docIDs[0].String(), docs[0].ID().String())
	require.Equal(t, docIDs[1].String(), docs[1].ID().String())
	require.Equal(t, []string{docIDs[0].String(), docIDs[1].String()}, documentIDs(docs))
}

func TestSetDocumentIDsFromJSONReturnsError(t *testing.T) {
	docID := client.NewDocIDV0(blocks.NewBlock([]byte("doc")).Cid())
	rawDocIDs, err := json.Marshal([]string{docID.String()})
	require.NoError(t, err)

	tests := []struct {
		name string
		docs []*client.Document
		data []byte
	}{
		{
			name: "invalid json",
			docs: []*client.Document{{}},
			data: []byte(`[`),
		},
		{
			name: "wrong count",
			docs: []*client.Document{{}, {}},
			data: rawDocIDs,
		},
		{
			name: "invalid doc id",
			docs: []*client.Document{{}},
			data: []byte(`["not-a-doc-id"]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, setDocumentIDsFromJSON(test.docs, test.data))
		})
	}
}
