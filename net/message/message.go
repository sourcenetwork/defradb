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
	"context"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/sourcenetwork/defradb/errors"
)

const messageVersion = "/defradb/0.0.1"

// Message is the interface that protocol messages need to implement
// in order to be compatible with [Send] and [Receive].
//
// If messages embed the [MetaData] type they will automatically satisfy the interface.
type Message interface {
	GetVersion() string
	SetVersion()
	GetMessageID() string
	SetMessageID(id string)
	GetPeerID() string
	SetPeerID(id string)
	GetPubkey() []byte
	SetPubkey(key []byte)
	GetSignature() []byte
	SetSignature(signature []byte)
	GetErrMessage() string
	SetErrMessage(err string)
}

// proto is the minum set of methods that protocols should implement to handle
// sending and receiving messages adequately.
type proto interface {
	Host() host.Host
	SetResponseChan(messageID string, message chan Message)
	DeleteResponseChan(messageID string)
	GetResponseChan(messageID string) (chan Message, bool)
}

// Reveive takes in a network stream and store the unmarshalled message in the provided [Message]
func Receive(s network.Stream, proto proto, m Message) error {
	b, err := io.ReadAll(s)
	if err != nil {
		resetErr := s.Reset()
		return errors.Join(err, resetErr)
	}
	err = s.Close()
	if err != nil {
		return err
	}

	err = cbor.Unmarshal(b, m)
	if err != nil {
		return err
	}

	err = verifyMessage(m)
	if err != nil {
		return err
	}

	messageChan, ok := proto.GetResponseChan(m.GetMessageID())
	if ok {
		messageChan <- m
		proto.DeleteResponseChan(m.GetMessageID())
	}

	return nil
}

// Send creates a new network stream with the provided peer, signs and set the appropriate meta data
// on the message and writes it to the stream.
//
// If wait is set to true and the response is never received, the call to this function will hang forever. It is the
// responsibility of the caller to set a reasonable timeout.
func Send(
	ctx context.Context,
	proto proto,
	m Message,
	pid peer.ID,
	protoID protocol.ID,
	wait bool,
) (resp Message, err error) {
	err = signAndSetMetaData(proto.Host(), m)
	if err != nil {
		return nil, err
	}
	signed, err := cbor.Marshal(m)
	if err != nil {
		return nil, err
	}

	s, err := proto.Host().NewStream(ctx, pid, protoID)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := s.Close()
		err = errors.Join(err, closeErr)
	}()

	var responseChan chan Message
	if wait {
		// we create a response channel before sending the message so that we can catch
		// the response.
		responseChan = make(chan Message, 1)
		proto.SetResponseChan(m.GetMessageID(), responseChan)
	}

	_, err = s.Write(signed)
	if err != nil {
		close(responseChan)
		proto.DeleteResponseChan(m.GetMessageID())
		resetErr := s.Reset()
		return nil, errors.Join(err, resetErr)
	}
	err = s.Close()
	if err != nil {
		return nil, err
	}

	if wait {
		select {
		case m := <-responseChan:
			if m.GetErrMessage() != "" {
				return nil, errors.New(m.GetErrMessage())
			}
			return m, nil
		case <-ctx.Done():
			close(responseChan)
			proto.DeleteResponseChan(m.GetMessageID())
			return nil, ErrResponseTimeout
		}
	}
	return nil, nil
}

func verifyMessage(m Message) error {
	pubkey, err := crypto.UnmarshalPublicKey(m.GetPubkey())
	if err != nil {
		return err
	}
	idFromKey, err := peer.IDFromPublicKey(pubkey)
	if err != nil {
		return err
	}
	peerID, err := peer.Decode(m.GetPeerID())
	if err != nil {
		return err
	}
	if idFromKey != peerID {
		return ErrPubkeyPeerIDMismatch
	}

	signature := m.GetSignature()
	// To verify the signed message, the signature itself has to be removed from the struct.
	m.SetSignature(nil)
	b, err := cbor.Marshal(m)
	if err != nil {
		return err
	}

	valid, err := pubkey.Verify(b, signature)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidSignature
	}
	return nil
}

func signAndSetMetaData(h host.Host, m Message) error {
	// if the message ID is already set, its because the message is a response message.
	if m.GetMessageID() == "" {
		messageID, err := uuid.NewV4()
		if err != nil {
			return err
		}
		m.SetMessageID(messageID.String())
	}

	nodePubKey, err := crypto.MarshalPublicKey(h.Peerstore().PubKey(h.ID()))
	if err != nil {
		return err
	}
	m.SetVersion()
	m.SetPubkey(nodePubKey)
	m.SetPeerID(h.ID().String())

	signature, err := signMessage(h, m)
	if err != nil {
		return err
	}
	m.SetSignature(signature)
	return nil
}

func signMessage(h host.Host, m Message) ([]byte, error) {
	b, err := cbor.Marshal(m)
	if err != nil {
		return nil, err
	}
	key := h.Peerstore().PrivKey(h.ID())
	return key.Sign(b)
}
