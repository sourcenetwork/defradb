// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"fmt"

	libpeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/net"
	pb "github.com/sourcenetwork/defradb/net/pb"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	grpcpeer "google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
)

const pubsubTopic = "encryption"

type PubSubServer interface {
	AddPubSubTopic(string, rpc.MessageHandler) error
	SendPubSubMessage(context.Context, string, []byte) (<-chan rpc.Response, error)
}

type pubSubService struct {
	ctx             context.Context
	peerID          libpeer.ID
	pubsub          PubSubServer
	keyRequestedSub *event.Subscription
	encstore        datastore.DSReaderWriter
	eventBus        *event.Bus
}

var _ Service = (*pubSubService)(nil)

func (s *pubSubService) GetKeys(ctx context.Context, keys ...core.EncStoreDocKey) (*encryption.Results, error) {
	res, ch := encryption.NewResults()

	err := s.requestEncryptionKey(ctx, keys, ch)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewPubSubService(
	ctx context.Context,
	peerID libpeer.ID,
	pubsub PubSubServer,
	eventBus *event.Bus,
	encstore datastore.DSReaderWriter,
) (*pubSubService, error) {
	s := &pubSubService{
		ctx:      ctx,
		peerID:   peerID,
		pubsub:   pubsub,
		encstore: encstore,
		eventBus: eventBus,
	}
	err := pubsub.AddPubSubTopic(pubsubTopic, s.handleRequestFromPeer)
	if err != nil {
		return nil, err
	}
	s.keyRequestedSub, err = eventBus.Subscribe(encryption.RequestKeysEventName)
	if err != nil {
		return nil, err
	}
	go s.handleKeyRequestedEvent()
	return s, nil
}

func (s *pubSubService) handleKeyRequestedEvent() {
	for {
		msg, isOpen := <-s.keyRequestedSub.Message()
		if !isOpen {
			return
		}

		if keyReqEvent, ok := msg.Data.(encryption.RequestKeysEvent); ok {
			go func() {
				results, err := s.GetKeys(s.ctx, keyReqEvent.Keys...)
				if err != nil {
					log.ErrorContextE(s.ctx, "Failed to get encryption keys", err)
				}

				encResult := <-results.Get()

				_, encryptor := encryption.ContextWithStore(s.ctx, s.encstore)

				for _, encItem := range encResult.Items {
					encKey, err := encryptor.GetKey(encItem.StoreKey)
					if err != nil {
						fmt.Printf(">>>>>>>>>> Failed to get stored key %s\n", encItem.StoreKey.ToString())
						continue
					}
					if len(encKey) == 0 {
						fmt.Printf(">>>>>>>>>> No key stored %s\n", encItem.StoreKey.ToString())
					} else {
						fmt.Printf(">>>>>>>>>> Has already key stored %s\n", encItem.StoreKey.ToString())
					}
					err = encryptor.SaveKey(encItem.StoreKey, encItem.EncryptionKey)
					if err != nil {
						log.ErrorContextE(s.ctx, "Failed to save encryption key", err)
						return
					}
					fmt.Printf(">>>>>>>>>> Saved key %s\n", encItem.StoreKey.ToString())
				}

				m := make(map[core.EncStoreDocKey][]byte)
				for _, item := range encResult.Items {
					m[item.StoreKey] = item.EncryptionKey
				}
				keyReqEvent.Resp <- encResult
				close(keyReqEvent.Resp)
			}()
		} else {
			log.ErrorContext(s.ctx, "Failed to cast event data to RequestKeysEvent")
		}
	}
}

// handleEncryptionMessage handles incoming FetchEncryptionKeyRequest messages from the pubsub network.
func (s *pubSubService) handleRequestFromPeer(peerID libpeer.ID, topic string, msg []byte) ([]byte, error) {
	// TODO: check how it makes sense and how much effort to separate net package so that it has
	// client-related and server-related code independently. Conceptually, they should no depend on each other.
	// Any common functionality (like hosting peer) can be shared.
	// This way we could make kms package depend only on client. The server would depend on kms.
	req := new(pb.FetchEncryptionKeyRequest)
	if err := proto.Unmarshal(msg, req); err != nil {
		log.ErrorContextE(s.ctx, "Failed to unmarshal pubsub message %s", err)
		return nil, err
	}

	// TODO: check if this NewGRPCPeer can be abstracted away or copied in this package.
	ctx := grpcpeer.NewContext(s.ctx, net.NewGRPCPeer(peerID))
	res, err := s.TryGenEncryptionKey(ctx, req)
	if err != nil {
		log.ErrorContextE(s.ctx, "failed attempt to get encryption key", err)
		return nil, errors.Wrap("failed attempt to get encryption key", err)
	}
	return res.MarshalVT()
}

func (s *pubSubService) prepareFetchEncryptionKeyRequest(
	encStoreKeys []core.EncStoreDocKey,
	ephemeralPublicKey []byte,
) (*pb.FetchEncryptionKeyRequest, error) {
	req := &pb.FetchEncryptionKeyRequest{
		EphemeralPublicKey: ephemeralPublicKey,
	}

	for _, encStoreKey := range encStoreKeys {
		encKey := &pb.EncryptionKeyTarget{
			DocID: []byte(encStoreKey.DocID),
			KeyID: []byte(encStoreKey.KeyID),
		}
		if encStoreKey.FieldName.HasValue() {
			encKey.FieldName = encStoreKey.FieldName.Value()
		}
		req.Targets = append(req.Targets, encKey)
	}

	return req, nil
}

// requestEncryptionKey publishes the given FetchEncryptionKeyRequest object on the PubSub network
func (s *pubSubService) requestEncryptionKey(
	ctx context.Context,
	encStoreKeys []core.EncStoreDocKey,
	result chan<- encryption.Result,
) error {
	ephPrivKey, err := crypto.GenerateX25519()
	if err != nil {
		return err
	}

	ephPubKeyBytes := ephPrivKey.PublicKey().Bytes()
	req, err := s.prepareFetchEncryptionKeyRequest(encStoreKeys, ephPubKeyBytes)
	if err != nil {
		return err
	}

	data, err := req.MarshalVT()
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	respChan, err := s.pubsub.SendPubSubMessage(ctx, pubsubTopic, data)
	//respChan, err := s.topic.Publish(ctx, data)
	if err != nil {
		return errors.Wrap("failed publishing to encryption thread", err)
	}

	//s.server.RequestEncryptionKey(ctx, req, ephPrivKey)

	go func() {
		s.handleFetchEncryptionKeyResponse(<-respChan, req, ephPrivKey, result)
	}()

	return nil
}

// handleFetchEncryptionKeyResponse handles incoming FetchEncryptionKeyResponse messages
func (s *pubSubService) handleFetchEncryptionKeyResponse(
	resp rpc.Response,
	req *pb.FetchEncryptionKeyRequest,
	privateKey *ecdh.PrivateKey,
	result chan<- encryption.Result,
) {
	defer close(result)

	var keyResp pb.FetchEncryptionKeyReply
	if err := proto.Unmarshal(resp.Data, &keyResp); err != nil {
		log.ErrorContextE(s.ctx, "Failed to unmarshal encryption key response", err)
		result <- encryption.Result{Error: err}
		return
	}

	decryptedData, err := crypto.DecryptECIES(
		keyResp.EncryptedKeys,
		privateKey,
		makeAssociatedData(req, resp.From),
	)

	if err != nil {
		log.ErrorContextE(s.ctx, "Failed to decrypt encryption key", err)
		result <- encryption.Result{Error: err}
		return
	}

	if len(decryptedData) != crypto.AESKeySize*len(keyResp.Targets) {
		log.ErrorContext(s.ctx, "Invalid decrypted data length")
		result <- encryption.Result{Error: errors.New("invalid decrypted data length")}
		return
	}

	resultEncItems := make([]encryption.Item, 0, len(keyResp.Targets))
	for _, target := range keyResp.Targets {
		optFieldName := immutable.None[string]()
		if target.FieldName != "" {
			optFieldName = immutable.Some(target.FieldName)
		}

		encKey := decryptedData[:crypto.AESKeySize]
		decryptedData = decryptedData[crypto.AESKeySize:]

		resultEncItems = append(resultEncItems, encryption.Item{
			StoreKey:      core.NewEncStoreDocKey(string(target.DocID), optFieldName, string(target.KeyID)),
			EncryptionKey: encKey,
		})
	}

	result <- encryption.Result{
		Items: resultEncItems,
	}
}

// makeAssociatedData creates the associated data for the encryption key request
func makeAssociatedData(req *pb.FetchEncryptionKeyRequest, peerID libpeer.ID) []byte {
	return encodeToBase64(bytes.Join([][]byte{
		req.EphemeralPublicKey,
		[]byte(peerID),
	}, []byte{}))
}

func (s *pubSubService) TryGenEncryptionKey(
	ctx context.Context,
	req *pb.FetchEncryptionKeyRequest,
) (*pb.FetchEncryptionKeyReply, error) {
	encryptionKeys, targets, err := s.getEncryptionKeys(ctx, req)
	if err != nil || len(encryptionKeys) == 0 {
		return nil, err
	}

	reqEphPubKey, err := crypto.X25519PublicKeyFromBytes(req.EphemeralPublicKey)
	if err != nil {
		return nil, errors.Wrap("failed to unmarshal ephemeral public key", err)
	}

	encryptedKey, err := crypto.EncryptECIES(encryptionKeys, reqEphPubKey, makeAssociatedData(req, s.peerID))
	if err != nil {
		return nil, errors.Wrap("failed to encrypt key for requester", err)
	}

	res := &pb.FetchEncryptionKeyReply{
		ReqEphemeralPublicKey: req.EphemeralPublicKey,
		Targets:               targets,
		EncryptedKeys:         encryptedKey,
	}

	return res, nil
}

// getEncryptionKeys retrieves the encryption keys for the given targets.
// It returns the encryption keys and the targets for which the keys were found.
func (s *pubSubService) getEncryptionKeys(
	ctx context.Context,
	req *pb.FetchEncryptionKeyRequest,
) ([]byte, []*pb.EncryptionKeyTarget, error) {
	encryptionKeys := make([]byte, 0, len(req.Targets))
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
		_, encryptor := encryption.ContextWithStore(ctx, s.encstore)
		if encryptor == nil {
			return nil, nil, encryption.ErrContextHasNoEncryptor
		}
		encKey, err := encryptor.GetKey(
			core.NewEncStoreDocKey(docID.String(), optFieldName, string(target.KeyID)),
		)
		if err != nil {
			return nil, nil, err
		}
		// TODO: we should test it somehow. For this this one peer should have some keys and
		// another one should have the others. https://github.com/sourcenetwork/defradb/issues/2895
		if len(encKey) == 0 {
			continue
		}
		targets = append(targets, target)
		encryptionKeys = append(encryptionKeys, encKey...)
	}
	return encryptionKeys, targets, nil
}

func encodeToBase64(data []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	return encoded
}
