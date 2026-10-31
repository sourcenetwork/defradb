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

type fakeHost struct{ client.Host }

func (fakeHost) ID() string                                         { return "sender" }
func (fakeHost) Pubkey() ([]byte, error)                            { return []byte{1}, nil }
func (fakeHost) Sign([]byte) ([]byte, error)                        { return []byte{2}, nil }
func (fakeHost) Send(context.Context, []byte, string, string) error { return nil }

// nackingProto answers every request with a reply that carries an error
// string, the way a saturated receiver nacks a PushLog.
type nackingProto struct {
	host client.Host
	nack string
}

func (p *nackingProto) Host() client.Host { return p.host }
func (p *nackingProto) SetResponseChan(messageID string, ch chan Message) {
	reply := &MetaData{}
	reply.SetMessageID(messageID)
	reply.SetErrMessage(p.nack)
	ch <- reply
}
func (p *nackingProto) DeleteResponseChan(string)                   {}
func (p *nackingProto) GetResponseChan(string) (chan Message, bool) { return nil, false }

// A receiver's nack must surface as Send's error; reading the request's own
// (empty) error string instead reports the rejected push as delivered, and the
// replicator then never writes a retry record for it.
func TestSend_ReturnsTheReplysErrMessage(t *testing.T) {
	proto := &nackingProto{host: fakeHost{}, nack: "at capacity: receiver is saturated, back off"}
	_, err := Send[*MetaData](context.Background(), proto, &MetaData{}, "receiver", "/defradb/rep_req/0.0.1")
	require.ErrorIs(t, err, ErrPeerRejected)
	require.ErrorContains(t, err, proto.nack)
}

// capturingProto keeps the response channel Send registered, standing in for an
// inbound handler goroutine that has already read the channel out of the map
// (message.go:127) and is about to deliver its reply into it.
type capturingProto struct {
	host client.Host
	ch   chan Message
}

func (p *capturingProto) Host() client.Host                           { return p.host }
func (p *capturingProto) SetResponseChan(_ string, ch chan Message)   { p.ch = ch }
func (p *capturingProto) DeleteResponseChan(string)                   {}
func (p *capturingProto) GetResponseChan(string) (chan Message, bool) { return nil, false }

// A reply that arrives after Send has given up must not land on a closed
// channel: Receive sends into a channel it read from the map before the
// timeout removed the entry, so closing on the timeout path is a send on a
// closed channel, which is a data race under -race and a panic without it.
func TestSend_LateReplyAfterTimeoutMustNotPanic(t *testing.T) {
	proto := &capturingProto{host: fakeHost{}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := Send[*MetaData](ctx, proto, &MetaData{}, "receiver", "/defradb/rep_req/0.0.1")
	require.ErrorIs(t, err, ErrResponseTimeout)

	require.NotPanics(t, func() { proto.ch <- &MetaData{} })
}

// keyHost signs with a real key, so messages it signs pass verifyMessage.
type keyHost struct {
	client.Host
	key crypto.PrivKey
}

func (h keyHost) ID() string {
	id, _ := peer.IDFromPrivateKey(h.key)
	return id.String()
}
func (h keyHost) Pubkey() ([]byte, error)       { return crypto.MarshalPublicKey(h.key.GetPublic()) }
func (h keyHost) Sign(b []byte) ([]byte, error) { return h.key.Sign(b) }

// fullChanProto hands Receive a response channel whose one slot is already taken,
// as when a duplicate reply arrives before the first one is consumed.
type fullChanProto struct{ ch chan Message }

func (p *fullChanProto) Host() client.Host                           { return nil }
func (p *fullChanProto) SetResponseChan(string, chan Message)        {}
func (p *fullChanProto) DeleteResponseChan(string)                   {}
func (p *fullChanProto) GetResponseChan(string) (chan Message, bool) { return p.ch, true }

// A duplicate reply must be dropped rather than block Receive: a blocked Receive
// never returns, so its deferred stream close never runs either.
func TestReceive_DuplicateReplyDoesNotBlock(t *testing.T) {
	key, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)
	host := keyHost{key: key}
	reply := &MetaData{}
	reply.SetMessageID("some message ID")
	require.NoError(t, signAndSetMetaData(host, reply))
	b, err := cbor.Marshal(reply)
	require.NoError(t, err)

	proto := &fullChanProto{ch: make(chan Message, 1)}
	proto.ch <- &MetaData{}

	done := make(chan error, 1)
	go func() { done <- Receive(bytes.NewReader(b), host.ID(), proto, &MetaData{}) }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Receive blocked delivering a duplicate reply")
	}
}
