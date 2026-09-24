// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package message

import (
	"bytes"
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

func TestReceive_StreamLargerThanMax_ReturnsErrMessageTooLarge(t *testing.T) {
	stream := bytes.NewReader(make([]byte, maxMessageSize+1))
	err := Receive(stream, "some peer ID", nil, &MetaData{})
	require.ErrorIs(t, err, ErrMessageTooLarge)
}

type closeRecordingStream struct {
	*bytes.Reader
	closed bool
}

func (s *closeRecordingStream) Close() error {
	s.closed = true
	return nil
}

func TestReceive_ClosesStream(t *testing.T) {
	stream := &closeRecordingStream{Reader: bytes.NewReader([]byte("not cbor"))}
	err := Receive(stream, "some peer ID", nil, &MetaData{})
	require.Error(t, err)
	require.True(t, stream.closed)
}

// testHost signs with a real key, so messages it signs pass Receive's verification.
type testHost struct {
	client.Host
	key crypto.PrivKey
	id  string
}

func newTestHost(t *testing.T) testHost {
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(key)
	require.NoError(t, err)
	return testHost{key: key, id: id.String()}
}

func (h testHost) ID() string                                       { return h.id }
func (h testHost) Pubkey() ([]byte, error)                          { return crypto.MarshalPublicKey(h.key.GetPublic()) }
func (h testHost) Sign(data []byte) ([]byte, error)                 { return h.key.Sign(data) }
func (testHost) Send(context.Context, []byte, string, string) error { return nil }

// testProto holds a single response channel: the one Send registers, or one a test sets.
type testProto struct {
	host testHost
	ch   chan Message
}

func (p *testProto) Host() client.Host { return p.host }

func (p *testProto) SetResponseChan(_ string, ch chan Message) { p.ch = ch }

func (p *testProto) DeleteResponseChan(string) {}

func (p *testProto) GetResponseChan(string) (chan Message, bool) { return p.ch, p.ch != nil }

func TestReceive_FullResponseChan_DoesNotBlock(t *testing.T) {
	host := newTestHost(t)
	reply := &MetaData{}
	require.NoError(t, signAndSetMetaData(host, reply))
	data, err := cbor.Marshal(reply)
	require.NoError(t, err)

	proto := &testProto{host: host, ch: make(chan Message, 1)}
	proto.ch <- &MetaData{}

	done := make(chan error, 1)
	go func() { done <- Receive(bytes.NewReader(data), host.ID(), proto, &MetaData{}) }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Receive blocked on a full response channel")
	}
}

func TestSend_ReplyAfterTimeout_DoesNotPanic(t *testing.T) {
	proto := &testProto{host: newTestHost(t)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Send[*MetaData](ctx, proto, &MetaData{}, "some peer ID", "/test/0.0.1")
	require.ErrorIs(t, err, ErrResponseTimeout)
	require.NotPanics(t, func() { proto.ch <- &MetaData{} })
}
