// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// docSyncHost is connected to one peer, which answers a document sync at once with no documents.
type docSyncHost struct {
	client.Host
	peer peer.ID
}

func (h *docSyncHost) ActivePeers() ([]string, error) {
	return []string{"/ip4/127.0.0.1/tcp/9171/p2p/" + h.peer.String()}, nil
}

func (h *docSyncHost) PublishToTopic(context.Context, string, []byte, bool) (<-chan client.PubsubResponse, error) {
	reply, err := cbor.Marshal(docSyncReply{Sender: h.peer.String()})
	if err != nil {
		return nil, err
	}
	responses := make(chan client.PubsubResponse, 1)
	responses <- client.PubsubResponse{From: h.peer.String(), Data: reply}
	return responses, nil
}

// A document sync returns once every connected peer has answered, rather than waiting out its
// deadline and reporting a timeout.
func TestSyncDocumentsReturnsOnceEveryPeerAnswers(t *testing.T) {
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(key)
	require.NoError(t, err)
	p := &P2P{host: &docSyncHost{peer: id}}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = p.syncDocuments(ctx, "collection", []string{"bae-00000000-0000-0000-0000-000000000000"})

	require.NoError(t, err)
}
