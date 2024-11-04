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

	"github.com/fxamacker/cbor/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	libpeer "github.com/libp2p/go-libp2p/core/peer"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

const pubsubTopic = "encryption"

type PubSubServer interface {
	AddPubSubTopic(string, rpc.MessageHandler) error
	SendPubSubMessage(context.Context, string, []byte) (<-chan rpc.Response, error)
}

type CollectionRetriever interface {
	RetrieveCollectionFromDocID(context.Context, string) (client.Collection, error)
}

type pubSubService struct {
	ctx             context.Context
	peerID          libpeer.ID
	pubsub          PubSubServer
	keyRequestedSub *event.Subscription
	eventBus        *event.Bus
	encStore        *ipldEncStorage
	acp             immutable.Option[acp.ACP]
	colRetriever    CollectionRetriever
	nodeDID         string
}

var _ Service = (*pubSubService)(nil)

func (s *pubSubService) GetKeys(ctx context.Context, cids ...cidlink.Link) (*encryption.Results, error) {
	res, ch := encryption.NewResults()

	err := s.requestEncryptionKeyFromPeers(ctx, cids, ch)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// NewPubSubService creates a new instance of the KMS service that is connected to the given PubSubServer,
// event bus and encryption storage.
//
// The service will subscribe to the "encryption" topic on the PubSubServer and to the
// "enc-keys-request" event on the event bus.
func NewPubSubService(
	ctx context.Context,
	peerID libpeer.ID,
	pubsub PubSubServer,
	eventBus *event.Bus,
	encstore datastore.Blockstore,
	acp immutable.Option[acp.ACP],
	colRetriever CollectionRetriever,
	nodeDID string,
) (*pubSubService, error) {
	s := &pubSubService{
		ctx:          ctx,
		peerID:       peerID,
		pubsub:       pubsub,
		eventBus:     eventBus,
		encStore:     newIPLDEncryptionStorage(encstore),
		acp:          acp,
		colRetriever: colRetriever,
		nodeDID:      nodeDID,
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

				defer close(keyReqEvent.Resp)

				select {
				case <-s.ctx.Done():
					return
				case encResult := <-results.Get():
					for _, encItem := range encResult.Items {
						_, err = s.encStore.put(s.ctx, encItem.Block)
						if err != nil {
							log.ErrorContextE(s.ctx, "Failed to save encryption key", err)
							return
						}
					}

					keyReqEvent.Resp <- encResult
				}
			}()
		} else {
			log.ErrorContext(s.ctx, "Failed to cast event data to RequestKeysEvent")
		}
	}
}

type fetchEncryptionKeyRequest struct {
	Identity           []byte
	Links              [][]byte
	EphemeralPublicKey []byte
}

// handleEncryptionMessage handles incoming FetchEncryptionKeyRequest messages from the pubsub network.
func (s *pubSubService) handleRequestFromPeer(peerID libpeer.ID, topic string, msg []byte) ([]byte, error) {
	req := new(fetchEncryptionKeyRequest)
	if err := cbor.Unmarshal(msg, req); err != nil {
		log.ErrorContextE(s.ctx, "Failed to unmarshal pubsub message %s", err)
		return nil, err
	}

	ctx := grpcpeer.NewContext(s.ctx, newGRPCPeer(peerID))
	res, err := s.tryGenEncryptionKeyLocally(ctx, req)
	if err != nil {
		log.ErrorContextE(s.ctx, "failed attempt to get encryption key", err)
		return nil, errors.Wrap("failed attempt to get encryption key", err)
	}
	return cbor.Marshal(res)
}

func (s *pubSubService) prepareFetchEncryptionKeyRequest(
	cids []cidlink.Link,
	ephemeralPublicKey []byte,
) (*fetchEncryptionKeyRequest, error) {
	req := &fetchEncryptionKeyRequest{
		Identity:           []byte(s.nodeDID),
		EphemeralPublicKey: ephemeralPublicKey,
	}

	req.Links = make([][]byte, len(cids))
	for i, cid := range cids {
		req.Links[i] = cid.Bytes()
	}

	return req, nil
}

// requestEncryptionKeyFromPeers publishes the given FetchEncryptionKeyRequest object on the PubSub network
func (s *pubSubService) requestEncryptionKeyFromPeers(
	ctx context.Context,
	cids []cidlink.Link,
	result chan<- encryption.Result,
) error {
	ephPrivKey, err := crypto.GenerateX25519()
	if err != nil {
		return err
	}

	ephPubKeyBytes := ephPrivKey.PublicKey().Bytes()
	req, err := s.prepareFetchEncryptionKeyRequest(cids, ephPubKeyBytes)
	if err != nil {
		return err
	}

	data, err := cbor.Marshal(req)
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	respChan, err := s.pubsub.SendPubSubMessage(ctx, pubsubTopic, data)
	if err != nil {
		return errors.Wrap("failed publishing to encryption thread", err)
	}

	go func() {
		s.handleFetchEncryptionKeyResponse(<-respChan, req, ephPrivKey, result)
	}()

	return nil
}

type fetchEncryptionKeyReply struct {
	Links              [][]byte
	Blocks             [][]byte
	EphemeralPublicKey []byte
}

// handleFetchEncryptionKeyResponse handles incoming FetchEncryptionKeyResponse messages
func (s *pubSubService) handleFetchEncryptionKeyResponse(
	resp rpc.Response,
	req *fetchEncryptionKeyRequest,
	privateKey *ecdh.PrivateKey,
	result chan<- encryption.Result,
) {
	defer close(result)

	var keyResp fetchEncryptionKeyReply
	if err := cbor.Unmarshal(resp.Data, &keyResp); err != nil {
		log.ErrorContextE(s.ctx, "Failed to unmarshal encryption key response", err)
		result <- encryption.Result{Error: err}
		return
	}

	resultEncItems := make([]encryption.Item, 0, len(keyResp.Blocks))
	for i, block := range keyResp.Blocks {
		decryptedData, err := crypto.DecryptECIES(
			block,
			privateKey,
			crypto.WithAAD(makeAssociatedData(req, resp.From)),
			crypto.WithPubKeyBytes(keyResp.EphemeralPublicKey),
			crypto.WithPubKeyPrepended(false),
		)

		if err != nil {
			log.ErrorContextE(s.ctx, "Failed to decrypt encryption key", err)
			result <- encryption.Result{Error: err}
			return
		}

		resultEncItems = append(resultEncItems, encryption.Item{
			Link:  keyResp.Links[i],
			Block: decryptedData,
		})
	}

	result <- encryption.Result{
		Items: resultEncItems,
	}
}

// makeAssociatedData creates the associated data for the encryption key request
func makeAssociatedData(req *fetchEncryptionKeyRequest, peerID libpeer.ID) []byte {
	return encodeToBase64(bytes.Join([][]byte{
		req.EphemeralPublicKey,
		[]byte(peerID),
	}, []byte{}))
}

func (s *pubSubService) tryGenEncryptionKeyLocally(
	ctx context.Context,
	req *fetchEncryptionKeyRequest,
) (*fetchEncryptionKeyReply, error) {
	blocks, err := s.getEncryptionKeysLocally(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return &fetchEncryptionKeyReply{}, nil
	}

	reqEphPubKey, err := crypto.X25519PublicKeyFromBytes(req.EphemeralPublicKey)
	if err != nil {
		return nil, errors.Wrap("failed to unmarshal ephemeral public key", err)
	}

	privKey, err := crypto.GenerateX25519()
	if err != nil {
		return nil, err
	}

	res := &fetchEncryptionKeyReply{
		Links:              req.Links,
		EphemeralPublicKey: privKey.PublicKey().Bytes(),
	}

	res.Blocks = make([][]byte, 0, len(blocks))

	for _, block := range blocks {
		encryptedBlock, err := crypto.EncryptECIES(
			block,
			reqEphPubKey,
			crypto.WithAAD(makeAssociatedData(req, s.peerID)),
			crypto.WithPrivKey(privKey),
			crypto.WithPubKeyPrepended(false),
		)
		if err != nil {
			return nil, errors.Wrap("failed to encrypt key for requester", err)
		}

		res.Blocks = append(res.Blocks, encryptedBlock)
	}

	return res, nil
}

// getEncryptionKeys retrieves the encryption keys for the given targets.
// It returns the encryption keys and the targets for which the keys were found.
func (s *pubSubService) getEncryptionKeysLocally(
	ctx context.Context,
	req *fetchEncryptionKeyRequest,
) ([][]byte, error) {
	blocks := make([][]byte, 0, len(req.Links))
	for _, link := range req.Links {
		encBlock, err := s.encStore.get(ctx, link)
		if err != nil {
			return nil, err
		}
		// TODO: we should test it somehow. For this this one peer should have some keys and
		// another one should have the others. https://github.com/sourcenetwork/defradb/issues/2895
		if encBlock == nil {
			continue
		}

		hasPerm, err := s.doesIdentityHaveDocPermission(ctx, string(req.Identity), encBlock)
		if err != nil {
			return nil, err
		}
		if !hasPerm {
			continue
		}

		encBlockBytes, err := encBlock.Marshal()
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, encBlockBytes)
	}
	return blocks, nil
}

func (s *pubSubService) doesIdentityHaveDocPermission(
	ctx context.Context,
	actorIdentity string,
	entBlock *coreblock.Encryption,
) (bool, error) {
	if !s.acp.HasValue() {
		return true, nil
	}

	docID := string(entBlock.DocID)
	collection, err := s.colRetriever.RetrieveCollectionFromDocID(ctx, docID)
	if err != nil {
		return false, err
	}

	policy := collection.Definition().Description.Policy
	if !policy.HasValue() || policy.Value().ID == "" || policy.Value().ResourceName == "" {
		return true, nil
	}

	policyID, resourceName := policy.Value().ID, policy.Value().ResourceName

	isRegistered, err := s.acp.Value().IsDocRegistered(ctx, policyID, resourceName, docID)
	if err != nil {
		return false, err
	}

	if !isRegistered {
		// Unrestricted access as it is a public document.
		return true, nil
	}

	hasPerm, err := s.acp.Value().CheckDocAccess(ctx, acp.ReadPermission, actorIdentity, policyID, resourceName, docID)

	return hasPerm, err
}

func encodeToBase64(data []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	return encoded
}

func newGRPCPeer(peerID libpeer.ID) *grpcpeer.Peer {
	return &grpcpeer.Peer{
		Addr: addr{peerID},
	}
}

// addr implements net.Addr and holds a libp2p peer ID.
type addr struct{ id libpeer.ID }

// Network returns the name of the network that this address belongs to (libp2p).
func (a addr) Network() string { return "libp2p" }

// String returns the peer ID of this address in string form (B58-encoded).
func (a addr) String() string { return a.id.String() }
