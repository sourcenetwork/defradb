// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package protocol

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

// routingHost is a minimal client.Host that routes Send calls to the target host's registered
// stream handler, so two protocol instances can exchange messages in-process. It carries a real
// libp2p key so message signing/verification succeeds.
type routingHost struct {
	client.Host
	id       string
	priv     crypto.PrivKey
	pubBytes []byte
	handlers map[string]client.StreamHandler
	registry map[string]*routingHost
}

func newRoutingHost(t *testing.T, registry map[string]*routingHost) *routingHost {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	require.NoError(t, err)
	pid, err := peer.IDFromPublicKey(pub)
	require.NoError(t, err)
	pubBytes, err := crypto.MarshalPublicKey(pub)
	require.NoError(t, err)

	h := &routingHost{
		id:       pid.String(),
		priv:     priv,
		pubBytes: pubBytes,
		handlers: make(map[string]client.StreamHandler),
		registry: registry,
	}
	registry[h.id] = h
	return h
}

func (h *routingHost) ID() string                       { return h.id }
func (h *routingHost) Pubkey() ([]byte, error)          { return h.pubBytes, nil }
func (h *routingHost) Sign(data []byte) ([]byte, error) { return h.priv.Sign(data) }

func (h *routingHost) SetStreamHandler(protocolID string, handler client.StreamHandler) {
	h.handlers[protocolID] = handler
}

func (h *routingHost) Send(_ context.Context, data []byte, peerID string, protocolID string) error {
	target, ok := h.registry[peerID]
	if !ok {
		return errors.New("unknown peer")
	}
	handler, ok := target.handlers[protocolID]
	if !ok {
		return errors.New("no handler for protocol")
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	from := h.id
	// libp2p invokes stream handlers in their own goroutine; mirror that here.
	go handler(bytes.NewReader(cp), from)
	return nil
}

func TestBlockSyncProtocol_RequestStreamsBodyAndEncCIDs(t *testing.T) {
	registry := make(map[string]*routingHost)
	requesterHost := newRoutingHost(t, registry)
	senderHost := newRoutingHost(t, registry)

	body := []byte("the-car-payload-bytes")
	wantEncCIDs := [][]byte{[]byte("enc-cid-1"), []byte("enc-cid-2")}

	var gotReq BlockSyncRequest
	handler := func(_ context.Context, req BlockSyncRequest) ([][]byte, func(io.Writer) error, error) {
		gotReq = req
		return wantEncCIDs, func(w io.Writer) error {
			_, err := w.Write(body)
			return err
		}, nil
	}

	NewBlockSyncProtocol(senderHost, handler)
	requester := NewBlockSyncProtocol(requesterHost, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &BlockSyncRequest{DocID: "doc-1", CollectionID: "col-1", Root: []byte("root-cid"), Full: true}

	var gotEncCIDs [][]byte
	var gotBody []byte
	err := requester.RequestBlocks(ctx, senderHost.ID(), req, func(encCIDs [][]byte, car io.Reader) error {
		gotEncCIDs = encCIDs
		b, err := io.ReadAll(car)
		gotBody = b
		return err
	})
	require.NoError(t, err)
	require.Equal(t, body, gotBody)
	require.Equal(t, wantEncCIDs, gotEncCIDs)
	require.Equal(t, "doc-1", gotReq.DocID)
	require.True(t, gotReq.Full)
}

func TestBlockSyncProtocol_HandlerErrorPropagates(t *testing.T) {
	registry := make(map[string]*routingHost)
	requesterHost := newRoutingHost(t, registry)
	senderHost := newRoutingHost(t, registry)

	handler := func(_ context.Context, _ BlockSyncRequest) ([][]byte, func(io.Writer) error, error) {
		return nil, nil, errors.New("access denied")
	}
	NewBlockSyncProtocol(senderHost, handler)
	requester := NewBlockSyncProtocol(requesterHost, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ingestCalled := false
	err := requester.RequestBlocks(ctx, senderHost.ID(), &BlockSyncRequest{DocID: "d"},
		func(_ [][]byte, _ io.Reader) error {
			ingestCalled = true
			return nil
		})
	require.ErrorContains(t, err, "access denied")
	require.False(t, ingestCalled)
}
