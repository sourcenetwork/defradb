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

type testHost struct{ client.Host }

func (testHost) ID() string                                         { return "sender" }
func (testHost) Pubkey() ([]byte, error)                            { return []byte{1}, nil }
func (testHost) Sign([]byte) ([]byte, error)                        { return []byte{2}, nil }
func (testHost) Send(context.Context, []byte, string, string) error { return nil }

// testProto keeps the response channel Send registers, as Receive holds it once it has taken it
// from the proto.
type testProto struct {
	ch chan Message
}

func (p *testProto) Host() client.Host { return testHost{} }

func (p *testProto) SetResponseChan(_ string, ch chan Message) { p.ch = ch }

func (p *testProto) DeleteResponseChan(string) {}

func (p *testProto) GetResponseChan(string) (chan Message, bool) { return nil, false }

func TestSend_ReplyAfterTimeout_DoesNotPanic(t *testing.T) {
	proto := &testProto{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Send[*MetaData](ctx, proto, &MetaData{}, "some peer ID", "/test/0.0.1")
	require.ErrorIs(t, err, ErrResponseTimeout)
	require.NotPanics(t, func() { proto.ch <- &MetaData{} })
}
