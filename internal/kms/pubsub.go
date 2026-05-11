// Copyright 2025 Democratized Data Foundation
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
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/encryption"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

const pubsubTopic = "encryption"

type PubSubServer interface {
	AddPubSubTopic(
		topicName string,
		subscribe bool,
		handler client.PubsubMessageHandler,
	) error
	PublishToTopic(
		ctx context.Context,
		topic string,
		data []byte,
		withMultiResponse bool,
	) (<-chan client.PubsubResponse, error)
}

type CollectionRetriever interface {
	RetrieveCollectionFromDocID(
		context.Context,
		string,
		immutable.Option[identity.Identity],
	) (client.Collection, error)
}

type pubSubService struct {
	ctx      context.Context
	peerID   string
	pubsub   PubSubServer
	encStore *ipldEncStorage
	// nodeACP returns the current NAC state. It is a getter rather than a captured
	// value because NAC may be initialised AFTER the KMS service is constructed
	nodeACP      func() acpDB.NACInfo
	documentACP  immutable.Option[dac.DocumentACP]
	colRetriever CollectionRetriever
	// nodeIdentity is this node's own identity. Used to authorize node-internal
	// operations (e.g. NAC-gated collection lookups while answering KMS key
	// requests). It is NOT the requester's identity, that travels on the wire
	// in fetchEncryptionKeyRequest.Identity and is consulted for DAC.
	nodeIdentity immutable.Option[identity.Identity]
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
	peerID string,
	pubsub PubSubServer,
	encstore datastore.Blockstore,
	nodeACP func() acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	colRetriever CollectionRetriever,
	nodeIdentity immutable.Option[identity.Identity],
) (*pubSubService, error) {
	s := &pubSubService{
		ctx:          ctx,
		peerID:       peerID,
		pubsub:       pubsub,
		encStore:     newIPLDEncryptionStorage(encstore),
		nodeACP:      nodeACP,
		documentACP:  documentACP,
		colRetriever: colRetriever,
		nodeIdentity: nodeIdentity,
	}
	err := pubsub.AddPubSubTopic(pubsubTopic, true, s.handleRequestFromPeer)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type fetchEncryptionKeyRequest struct {
	Identity           []byte
	Links              [][]byte
	EphemeralPublicKey []byte
}

// handleEncryptionMessage handles incoming FetchEncryptionKeyRequest messages from the pubsub network.
func (s *pubSubService) handleRequestFromPeer(peerID string, topic string, msg []byte) ([]byte, error) {
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
	ctx context.Context,
	cids []cidlink.Link,
	ephemeralPublicKey []byte,
) (*fetchEncryptionKeyRequest, error) {
	// Prefer the caller's identity from ctx; fall back to the node identity
	// for paths that don't carry a user identity (e.g. gossip-triggered fetches).
	ident := iIdentity.FromContext(ctx)
	if !ident.HasValue() {
		ident = s.nodeIdentity
	}
	var didBytes []byte
	if ident.HasValue() {
		didBytes = []byte(ident.Value().DID())
	}
	req := &fetchEncryptionKeyRequest{
		Identity:           didBytes,
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
	req, err := s.prepareFetchEncryptionKeyRequest(ctx, cids, ephPubKeyBytes)
	if err != nil {
		return err
	}

	data, err := cbor.Marshal(req)
	if err != nil {
		return errors.Wrap("failed to marshal pubsub message", err)
	}

	respChan, err := s.pubsub.PublishToTopic(ctx, pubsubTopic, data, false)
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
	resp client.PubsubResponse,
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

		_, err = s.encStore.put(context.Background(), decryptedData)
		if err != nil {
			log.ErrorContextE(s.ctx, "Failed to store encryption key", err)
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
func makeAssociatedData(req *fetchEncryptionKeyRequest, peerID string) []byte {
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
	var actorIdentity immutable.Option[identity.Identity]
	if len(req.Identity) > 0 {
		actorIdentity = immutable.Some(identity.FromDID(string(req.Identity)))
	}

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

		docID := string(encBlock.DocID)
		if docID != "" {
			// Doc-scoped block: gate on per-doc DAC.
			hasPerm, err := s.doesIdentityHaveDocPermission(ctx, actorIdentity, docID)
			if err != nil {
				return nil, err
			}
			if !hasPerm {
				continue
			}
		} else {
			// Collection-scoped block (e.g. a `@branchable` collection's own head).
			// The block doesn't carry a CollectionID, so we can't run a per-collection
			// DAC check. Fall back to a node-level NAC gate: if the requester has no
			// authorized access on this node, refuse to serve. When NAC is not enabled
			// this is a no-op, preserving existing behaviour.
			hasNodeAccess, err := s.doesIdentityHaveNodeReadAccess(ctx, actorIdentity)
			if err != nil {
				return nil, err
			}
			if !hasNodeAccess {
				continue
			}
		}

		encBlockBytes, err := encBlock.Marshal()
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, encBlockBytes)
	}
	return blocks, nil
}

// doesIdentityHaveDocPermission asks whether actorIdentity may read docID.
// The collection lookup runs as the node itself (NAC), the DAC check runs as
// the requester. docID must be non-empty.
func (s *pubSubService) doesIdentityHaveDocPermission(
	ctx context.Context,
	actorIdentity immutable.Option[identity.Identity],
	docID string,
) (bool, error) {
	if !s.documentACP.HasValue() {
		return true, nil
	}

	collection, err := s.colRetriever.RetrieveCollectionFromDocID(ctx, docID, s.nodeIdentity)
	if err != nil {
		return false, err
	}

	return acpDB.CheckAccessOfDocOnCollectionWithACP(
		ctx,
		actorIdentity,
		s.nodeACP(),
		s.documentACP.Value(),
		collection,
		acpTypes.DocumentReadPerm,
		docID,
	)
}

// doesIdentityHaveNodeReadAccess returns true if actorIdentity is authorized to
// perform a read on this node, used as a fallback gate for encryption blocks
// that have no DocID (e.g. a `@branchable` collection's own head, where there
// is no per-doc ACL to consult). Returns true unconditionally when NAC is not
// enabled.
func (s *pubSubService) doesIdentityHaveNodeReadAccess(
	ctx context.Context,
	actorIdentity immutable.Option[identity.Identity],
) (bool, error) {
	var actorDID string
	if actorIdentity.HasValue() {
		actorDID = actorIdentity.Value().DID()
	}

	err := acpDB.CheckNodeOperationAccess(
		ctx,
		actorDID,
		s.nodeACP(),
		acpTypes.NodeReadDocumentPerm,
		acpTypes.NodeACPObject,
	)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, client.ErrNotAuthorizedToPerformOperation) {
		return false, nil
	}
	return false, err
}

func encodeToBase64(data []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	return encoded
}

func newGRPCPeer(peerID string) *grpcpeer.Peer {
	return &grpcpeer.Peer{
		Addr: addr{peerID},
	}
}

// addr implements net.Addr and holds a libp2p peer ID.
type addr struct{ id string }

// Network returns the name of the network that this address belongs to (libp2p).
func (a addr) Network() string { return "libp2p" }

// String returns the peer ID of this address in string form (B58-encoded).
func (a addr) String() string { return a.id }
