// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	libpeer "github.com/libp2p/go-libp2p/core/peer"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	grpcpeer "google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/encryption"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

const encryptionTopic = "encryption"

// getEncryptionKeys retrieves the encryption keys for the given targets.
// It returns the encryption keys and the targets for which the keys were found.
func (s *server) getEncryptionKeys(
	ctx context.Context,
	req *pb.FetchEncryptionKeyRequest,
) ([]byte, []*pb.EncryptionKeyTarget, error) {
	encryptionKeys := make([]byte, 0)
	targets := make([]*pb.EncryptionKeyTarget, 0, len(req.Targets))
	for _, target := range req.Targets {
		docID, err := client.NewDocIDFromString(string(target.DocID))
		if err != nil {
			return nil, nil, err
		}

		optFieldName := immutable.None[string]()
		if target.FieldName != "" {
			optFieldName = immutable.Some(target.FieldName)
		}
		encKey, err := encryption.GetKey(
			encryption.ContextWithStore(ctx, s.peer.encstore),
			core.NewEncStoreDocKey(docID.String(), optFieldName, target.Height),
		)
		if err != nil {
			return nil, nil, err
		}
		// TODO: we should test it somehow. For this this one peer should have some keys and
		// another one should have the others
		if len(encKey) == 0 {
			continue
		}
		targets = append(targets, target)
		encryptionKeys = append(encryptionKeys, encKey...)
	}
	return encryptionKeys, targets, nil
}

func (s *server) TryGenEncryptionKey(
	ctx context.Context,
	req *pb.FetchEncryptionKeyRequest,
) (*pb.FetchEncryptionKeyReply, error) {
	peerID, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reqPubKey := s.peer.host.Peerstore().PubKey(peerID)

	isValid, err := s.verifyRequestSignature(req, reqPubKey)
	if err != nil {
		return nil, errors.Wrap("invalid signature", err)
	}

	if !isValid {
		return nil, errors.New("invalid signature")
	}

	encryptionKeys, targets, err := s.getEncryptionKeys(ctx, req)
	if err != nil || len(encryptionKeys) == 0 {
		return nil, err
	}

	reqEphPubKey, err := crypto.X25519PublicKeyFromBytes(req.EphemeralPublicKey)
	if err != nil {
		return nil, errors.Wrap("failed to unmarshal ephemeral public key", err)
	}

	encryptedKey, err := crypto.EncryptECIES(encryptionKeys, reqEphPubKey, makeAssociatedData(req, s.peer.PeerID()))
	if err != nil {
		return nil, errors.Wrap("failed to encrypt key for requester", err)
	}

	res := &pb.FetchEncryptionKeyReply{
		SchemaRoot:            req.SchemaRoot,
		ReqEphemeralPublicKey: req.EphemeralPublicKey,
		Targets:               targets,
		EncryptedKeys:         encryptedKey,
	}

	res.Signature, err = s.signResponse(res)
	if err != nil {
		return nil, errors.Wrap("failed to sign response", err)
	}

	return res, nil
}

func (s *server) verifyRequestSignature(req *pb.FetchEncryptionKeyRequest, pubKey libp2pCrypto.PubKey) (bool, error) {
	return pubKey.Verify(hashFetchEncryptionKeyRequest(req), req.Signature)
}

func hashFetchEncryptionKeyReply(res *pb.FetchEncryptionKeyReply) []byte {
	hash := sha256.New()
	hash.Write(res.EncryptedKeys)
	hash.Write(res.SchemaRoot)
	hash.Write(res.ReqEphemeralPublicKey)
	for _, target := range res.Targets {
		hash.Write(target.DocID)
		hash.Write([]byte(target.FieldName))
		heightBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(heightBytes, target.Height)
		hash.Write(heightBytes)
	}
	return hash.Sum(nil)
}

func (s *server) signResponse(res *pb.FetchEncryptionKeyReply) ([]byte, error) {
	privKey := s.peer.host.Peerstore().PrivKey(s.peer.host.ID())
	return privKey.Sign(hashFetchEncryptionKeyReply(res))
}

// addPubSubEncryptionTopic subscribes to a topic on the pubsub network
func (s *server) addPubSubEncryptionTopic() error {
	if s.peer.ps == nil {
		return nil
	}

	t, err := rpc.NewTopic(s.peer.ctx, s.peer.ps, s.peer.host.ID(), encryptionTopic, true)
	if err != nil {
		return err
	}

	t.SetEventHandler(s.pubSubEventHandler)
	t.SetMessageHandler(s.pubSubEncryptionMessageHandler)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.topics[encryptionTopic] = pubsubTopic{
		Topic:      t,
		subscribed: true,
	}
	return nil
}

// pubSubEncryptionMessageHandler handles incoming FetchEncryptionKeyRequest messages from the pubsub network.
func (s *server) pubSubEncryptionMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
	req := new(pb.FetchEncryptionKeyRequest)
	if err := proto.Unmarshal(msg, req); err != nil {
		log.ErrorContextE(s.peer.ctx, "Failed to unmarshal pubsub message %s", err)
		return nil, err
	}

	ctx := grpcpeer.NewContext(s.peer.ctx, &grpcpeer.Peer{
		Addr: addr{from},
	})
	res, err := s.TryGenEncryptionKey(ctx, req)
	if err != nil {
		log.ErrorContextE(s.peer.ctx, "failed attempt to get encryption key", err)
		return nil, errors.Wrap("failed attempt to get encryption key", err)
	}
	return res.MarshalVT()
}

func (s *server) prepareFetchEncryptionKeyRequest(
	evt encryption.RequestKeysEvent,
	ephemeralPublicKey []byte,
) (*pb.FetchEncryptionKeyRequest, error) {
	req := &pb.FetchEncryptionKeyRequest{
		SchemaRoot:         []byte(evt.SchemaRoot),
		EphemeralPublicKey: ephemeralPublicKey,
	}

	for _, encStoreKey := range evt.Keys {
		encKey := &pb.EncryptionKeyTarget{
			DocID:  []byte(encStoreKey.DocID),
			Height: encStoreKey.BlockHeight,
		}
		if encStoreKey.FieldName.HasValue() {
			encKey.FieldName = encStoreKey.FieldName.Value()
		}
		req.Targets = append(req.Targets, encKey)
	}

	signature, err := s.signRequest(req)
	if err != nil {
		return nil, errors.Wrap("failed to sign request", err)
	}

	req.Signature = signature

	return req, nil
}

// requestEncryptionKey publishes the given FetchEncryptionKeyRequest object on the PubSub network
func (s *server) requestEncryptionKey(ctx context.Context, evt encryption.RequestKeysEvent) error {
	if s.peer.ps == nil { // skip if we aren't running with a pubsub net
		return nil
	}

	ephPrivKey, err := crypto.GenerateX25519()
	if err != nil {
		return err
	}

	ephPubKeyBytes := ephPrivKey.PublicKey().Bytes()
	req, err := s.prepareFetchEncryptionKeyRequest(evt, ephPubKeyBytes)
	if err != nil {
		return err
	}

	data, err := req.MarshalVT()
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	s.mu.Lock()
	t := s.topics[encryptionTopic]
	s.mu.Unlock()
	respChan, err := t.Publish(ctx, data)
	if err != nil {
		return errors.Wrap(fmt.Sprintf("failed publishing to thread %s", encryptionTopic), err)
	}

	s.sessions = append(s.sessions, newSession(string(ephPubKeyBytes), ephPrivKey))

	go func() {
		s.handleFetchEncryptionKeyResponse(<-respChan, req)
	}()

	return nil
}

func hashFetchEncryptionKeyRequest(req *pb.FetchEncryptionKeyRequest) []byte {
	hash := sha256.New()
	hash.Write(req.SchemaRoot)
	hash.Write(req.EphemeralPublicKey)
	for _, target := range req.Targets {
		hash.Write(target.DocID)
		hash.Write([]byte(target.FieldName))
		heightBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(heightBytes, target.Height)
		hash.Write(heightBytes)
	}
	return hash.Sum(nil)
}

func (s *server) signRequest(req *pb.FetchEncryptionKeyRequest) ([]byte, error) {
	privKey := s.peer.host.Peerstore().PrivKey(s.peer.host.ID())
	return privKey.Sign(hashFetchEncryptionKeyRequest(req))
}

// handleFetchEncryptionKeyResponse handles incoming FetchEncryptionKeyResponse messages
func (s *server) handleFetchEncryptionKeyResponse(resp rpc.Response, req *pb.FetchEncryptionKeyRequest) {
	var keyResp pb.FetchEncryptionKeyReply
	if err := proto.Unmarshal(resp.Data, &keyResp); err != nil {
		log.ErrorContextE(s.peer.ctx, "Failed to unmarshal encryption key response", err)
		return
	}

	isValid, err := s.verifyResponseSignature(&keyResp, resp.From)
	if err != nil {
		log.ErrorContextE(s.peer.ctx, "Failed to verify response signature", err)
		return
	}

	if !isValid {
		log.ErrorContext(s.peer.ctx, "Invalid response signature")
		return
	}

	session := s.extractSessionAndRemoveOldOnes(string(keyResp.ReqEphemeralPublicKey))
	if session == nil {
		log.ErrorContext(s.peer.ctx, "Failed to find session for ephemeral public key")
		return
	}

	decryptedData, err := crypto.DecryptECIES(
		keyResp.EncryptedKeys,
		session.privateKey,
		makeAssociatedData(req, resp.From),
	)

	if err != nil {
		log.ErrorContextE(s.peer.ctx, "Failed to decrypt encryption key", err)
		return
	}

	if len(decryptedData) != crypto.AESKeySize*len(keyResp.Targets) {
		log.ErrorContext(s.peer.ctx, "Invalid decrypted data length")
		return
	}

	eventData := make(map[core.EncStoreDocKey][]byte)
	for _, target := range keyResp.Targets {
		optFieldName := immutable.None[string]()
		if target.FieldName != "" {
			optFieldName = immutable.Some(target.FieldName)
		}

		encKey := decryptedData[:crypto.AESKeySize]
		decryptedData = decryptedData[crypto.AESKeySize:]

		eventData[core.NewEncStoreDocKey(string(target.DocID), optFieldName, target.Height)] = encKey
	}

	s.peer.bus.Publish(encryption.NewKeysRetrievedMessage(string(req.SchemaRoot), eventData))
}

// makeAssociatedData creates the associated data for the encryption key request
func makeAssociatedData(req *pb.FetchEncryptionKeyRequest, peerID libpeer.ID) []byte {
	return bytes.Join([][]byte{
		[]byte(req.SchemaRoot),
		[]byte(req.EphemeralPublicKey),
		[]byte(peerID),
	}, []byte{})
}

func (s *server) verifyResponseSignature(res *pb.FetchEncryptionKeyReply, fromPeer peer.ID) (bool, error) {
	pubKey := s.peer.host.Peerstore().PubKey(fromPeer)
	return pubKey.Verify(hashFetchEncryptionKeyReply(res), res.Signature)
}
