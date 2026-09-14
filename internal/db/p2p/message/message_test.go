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

	"github.com/sourcenetwork/defradb/client"

	"github.com/stretchr/testify/require"
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

// go-libp2p releases a stream's resource-manager reservation only when the
// local side closes it. Every inbound stream handler hands its stream to
// Receive and never touches it again, so Receive must close it — otherwise each
// accepted stream holds one of the per-(protocol, peer) inbound slots forever
// and the peer's 97th stream is refused.
func TestReceive_ClosesTheStreamWhenItCan(t *testing.T) {
	stream := &closeRecordingStream{Reader: bytes.NewReader([]byte("not cbor"))}
	_ = Receive(stream, "some peer ID", nil, &MetaData{})
	require.True(t, stream.closed, "inbound stream left open after Receive")
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
	require.EqualError(t, err, proto.nack)
}
