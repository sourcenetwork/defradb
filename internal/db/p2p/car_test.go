// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"bytes"
	"context"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	gocar "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/immutable"
)

// buildTestBlock constructs a minimal coreblock.Block, marshals it to bytes,
// and derives its CID using the standard DAG-CBOR link prototype.
func buildTestBlock(t *testing.T) (*coreblock.Block, cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:               []byte("testDoc"),
				Priority:            1,
				CollectionVersionID: "v1",
				Status:              1,
			},
		},
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := block.GenerateLink()
	require.NoError(t, err)

	return block, link.Cid, rawBytes
}

// buildExtraBlock constructs a non-root coreblock.Block (a field-level LWW block).
func buildExtraBlock(t *testing.T) (cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:               []byte("testDoc"),
				FieldName:           "name",
				Priority:            1,
				CollectionVersionID: "v1",
				Data:                []byte("Alice"),
			},
		},
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := block.GenerateLink()
	require.NoError(t, err)

	return link.Cid, rawBytes
}

func buildEncryptionBlock(t *testing.T) (cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Encryption{
		DocID: []byte("testDoc"),
		Key:   []byte("secret"),
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := coreblock.GetLinkFromNode(block.GenerateNode())
	require.NoError(t, err)

	return link.Cid, rawBytes
}

func buildSignatureBlock(t *testing.T) (cid.Cid, []byte) {
	t.Helper()
	return buildSignatureBlockForIdentity(t, "did:key:test", "signature")
}

func buildSignatureBlockForIdentity(t *testing.T, identity string, value string) (cid.Cid, []byte) {
	t.Helper()

	block := &coreblock.Signature{
		Header: coreblock.SignatureHeader{
			Type:     coreblock.SignatureTypeEd25519,
			Identity: []byte(identity),
		},
		Value: []byte(value),
	}

	rawBytes, err := block.Marshal()
	require.NoError(t, err)

	link, err := coreblock.GetLinkFromNode(block.GenerateNode())
	require.NoError(t, err)

	return link.Cid, rawBytes
}

func buildSignatureBlockForBlock(t *testing.T, block *coreblock.Block) (cid.Cid, []byte) {
	t.Helper()

	ident, err := identity.Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)

	blockBytes, err := block.Marshal()
	require.NoError(t, err)

	signatureBytes, err := ident.PrivateKey().Sign(blockBytes)
	require.NoError(t, err)

	signature := &coreblock.Signature{
		Header: coreblock.SignatureHeader{
			Type:     coreblock.SignatureTypeEd25519,
			Identity: []byte(ident.PublicKey().String()),
		},
		Value: signatureBytes,
	}

	rawBytes, err := signature.Marshal()
	require.NoError(t, err)

	link, err := coreblock.GetLinkFromNode(signature.GenerateNode())
	require.NoError(t, err)

	return link.Cid, rawBytes
}

// buildCARFromBlocks creates a CARv1 byte slice with the given root and blocks
// without relying on any db package.
func buildCARFromBlocks(t *testing.T, rootCID cid.Cid, blocks map[cid.Cid][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer, err := storage.NewWritable(&buf, []cid.Cid{rootCID}, gocar.WriteAsCarV1(true))
	require.NoError(t, err)

	for c, data := range blocks {
		link := cidlink.Link{Cid: c}
		err := writer.Put(t.Context(), link.Binary(), data)
		require.NoError(t, err)
	}

	return buf.Bytes()
}

type carTestDB struct {
	DB

	rootstore  corekv.TxnStore
	multistore *datastore.Multistore
}

func (db *carTestDB) Rootstore() corekv.TxnStore {
	return db.rootstore
}

func (db *carTestDB) Multistore() *datastore.Multistore {
	return db.multistore
}

func TestParseCAR_RoundTrip(t *testing.T) {
	_, rootCID, rootBytes := buildTestBlock(t)
	extraCID, extraBytes := buildExtraBlock(t)

	allBlocks := map[cid.Cid][]byte{
		rootCID:  rootBytes,
		extraCID: extraBytes,
	}

	carData := buildCARFromBlocks(t, rootCID, allBlocks)

	result, err := parseCAR(carData)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.rootBlock)

	// Verify the root block CID matches what we put in.
	rootLink, err := result.rootBlock.GenerateLink()
	require.NoError(t, err)
	require.Equal(t, rootCID, rootLink.Cid)

	// All regular blocks (root + extra) should be present.
	require.Len(t, result.regularBlocks, 2)
	require.Empty(t, result.encBlocks)
	require.Empty(t, result.sigBlocks)

	cidSet := make(map[cid.Cid]struct{})
	for _, blk := range result.regularBlocks {
		cidSet[blk.Cid()] = struct{}{}
	}
	require.Contains(t, cidSet, rootCID)
	require.Contains(t, cidSet, extraCID)
}

func TestParseCAR_SeparatesSidecarBlocks(t *testing.T) {
	_, rootCID, rootBytes := buildTestBlock(t)
	extraCID, extraBytes := buildExtraBlock(t)
	encCID, encBytes := buildEncryptionBlock(t)
	sigCID, sigBytes := buildSignatureBlock(t)

	allBlocks := map[cid.Cid][]byte{
		rootCID:  rootBytes,
		extraCID: extraBytes,
		encCID:   encBytes,
		sigCID:   sigBytes,
	}

	carData := buildCARFromBlocks(t, rootCID, allBlocks)

	result, err := parseCAR(carData)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.rootBlock)

	require.Len(t, result.regularBlocks, 2)
	require.Len(t, result.encBlocks, 1)
	require.Len(t, result.sigBlocks, 1)
	require.Equal(t, encCID, result.encBlocks[0].Cid())
	require.Equal(t, sigCID, result.sigBlocks[0].Cid())
}

func TestCollectSignatureBlocks_CollectsLinkedSignatureBlocks(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())

	_, rootCID, _ := buildTestBlock(t)
	existingSigCID, existingSigBytes := buildSignatureBlockForIdentity(t, "did:key:existing", "existing signature")
	newSigCID, newSigBytes := buildSignatureBlockForIdentity(t, "did:key:new", "new signature")

	putBlock := func(blockCID cid.Cid, data []byte) {
		t.Helper()

		rawBlock, err := blocks.NewBlockWithCid(data, blockCID)
		require.NoError(t, err)
		require.NoError(t, multistore.Blockstore().Put(ctx, rawBlock))
	}
	putBlock(existingSigCID, existingSigBytes)
	putBlock(newSigCID, newSigBytes)

	storeSignatureLink := func(sigCID cid.Cid, sigBytes []byte) {
		t.Helper()

		signature, err := coreblock.GetSignatureBlockFromBytes(sigBytes)
		require.NoError(t, err)
		require.NoError(
			t,
			coreblock.StoreSignatureLink(ctx, multistore.Systemstore(), rootCID, signature, cidlink.Link{Cid: sigCID}),
		)
	}
	storeSignatureLink(existingSigCID, existingSigBytes)
	storeSignatureLink(newSigCID, newSigBytes)

	existingBytes := []byte("already collected")
	collected := map[string]*collectedBlock{
		existingSigCID.String(): {
			cid:      existingSigCID,
			rawBytes: existingBytes,
		},
	}

	err := collectSignatureBlocks(ctx, multistore.Blockstore(), multistore.Systemstore(), rootCID, collected)
	require.NoError(t, err)

	require.Len(t, collected, 2)
	require.Equal(t, existingSigCID, collected[existingSigCID.String()].cid)
	require.Equal(t, existingBytes, collected[existingSigCID.String()].rawBytes)
	require.Equal(t, newSigCID, collected[newSigCID.String()].cid)
	require.Equal(t, newSigBytes, collected[newSigCID.String()].rawBytes)
}

func TestCollectSignatureBlocks_IfLinkedSignatureBlockMissing_ShouldError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())

	_, rootCID, _ := buildTestBlock(t)
	missingSigCID, missingSigBytes := buildSignatureBlockForIdentity(t, "did:key:missing", t.Name())
	signature, err := coreblock.GetSignatureBlockFromBytes(missingSigBytes)
	require.NoError(t, err)
	require.NoError(
		t,
		coreblock.StoreSignatureLink(ctx, multistore.Systemstore(), rootCID, signature, cidlink.Link{Cid: missingSigCID}),
	)

	collected := make(map[string]*collectedBlock)

	err = collectSignatureBlocks(ctx, multistore.Blockstore(), multistore.Systemstore(), rootCID, collected)
	require.Error(t, err)
	require.Empty(t, collected)
}

func TestGenerateCAR_IncludesSidecarSignatureBlocks(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	_, headCID, headBytes := buildTestBlock(t)
	rootBlock, _, _ := buildTestBlock(t)
	rootBlock.Delta.DocCompositeDelta.Priority = 2
	childCID, childBytes := buildExtraBlock(t)
	rootBlock.Heads = []cidlink.Link{{Cid: headCID}}
	rootBlock.Links = []coreblock.DAGLink{
		coreblock.NewDAGLink("name", cidlink.Link{Cid: childCID}),
	}
	rootBytes, err := rootBlock.Marshal()
	require.NoError(t, err)

	rootLink, err := rootBlock.GenerateLink()
	require.NoError(t, err)
	rootCID := rootLink.Cid

	putBlock := func(blockCID cid.Cid, data []byte) {
		t.Helper()

		rawBlock, err := blocks.NewBlockWithCid(data, blockCID)
		require.NoError(t, err)
		require.NoError(t, multistore.Blockstore().Put(ctx, rawBlock))
	}
	putBlock(headCID, headBytes)
	putBlock(rootCID, rootBytes)
	putBlock(childCID, childBytes)

	rootSigCID, rootSigBytes := buildSignatureBlockForIdentity(t, "did:key:root", "root signature")
	headSigCID, headSigBytes := buildSignatureBlockForIdentity(t, "did:key:head", "head signature")
	childSigCID, childSigBytes := buildSignatureBlockForIdentity(t, "did:key:child", "child signature")
	putBlock(rootSigCID, rootSigBytes)
	putBlock(headSigCID, headSigBytes)
	putBlock(childSigCID, childSigBytes)

	storeSignatureLink := func(blockCID cid.Cid, sigCID cid.Cid, sigBytes []byte) {
		t.Helper()

		sig, err := coreblock.GetSignatureBlockFromBytes(sigBytes)
		require.NoError(t, err)
		require.NoError(
			t,
			coreblock.StoreSignatureLink(ctx, multistore.Systemstore(), blockCID, sig, cidlink.Link{Cid: sigCID}),
		)
	}
	storeSignatureLink(rootCID, rootSigCID, rootSigBytes)
	storeSignatureLink(headCID, headSigCID, headSigBytes)
	storeSignatureLink(childCID, childSigCID, childSigBytes)

	carData, err := p.generateCARForBlocksWithBytes(ctx, []rootBlockWithBytes{{
		block:    rootBlock,
		rawBytes: rootBytes,
	}})
	require.NoError(t, err)

	result, err := parseCAR(carData)
	require.NoError(t, err)
	require.Len(t, result.regularBlocks, 3)
	require.Len(t, result.sigBlocks, 3)

	regularCIDs := make(map[cid.Cid]struct{}, len(result.regularBlocks))
	for _, block := range result.regularBlocks {
		regularCIDs[block.Cid()] = struct{}{}
	}
	require.Contains(t, regularCIDs, rootCID)
	require.Contains(t, regularCIDs, headCID)
	require.Contains(t, regularCIDs, childCID)

	sigCIDs := make(map[cid.Cid]struct{}, len(result.sigBlocks))
	for _, block := range result.sigBlocks {
		sigCIDs[block.Cid()] = struct{}{}
	}
	require.Contains(t, sigCIDs, rootSigCID)
	require.Contains(t, sigCIDs, headSigCID)
	require.Contains(t, sigCIDs, childSigCID)
}

func TestVerifyCARBlockSignatures_VerifiesSidecarSignatures(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, rootBytes := buildTestBlock(t)
	rootRawBlock, err := blocks.NewBlockWithCid(rootBytes, rootCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, rootRawBlock))

	sigCID, sigBytes := buildSignatureBlockForBlock(t, rootBlock)
	sigRawBlock, err := blocks.NewBlockWithCid(sigBytes, sigCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, sigRawBlock))

	signature, err := coreblock.GetSignatureBlockFromBytes(sigBytes)
	require.NoError(t, err)
	require.NoError(
		t,
		coreblock.StoreSignatureLink(ctx, multistore.Systemstore(), rootCID, signature, cidlink.Link{Cid: sigCID}),
	)

	require.NoError(t, p.verifyCARBlockSignatures(ctx, []blocks.Block{rootRawBlock}))
}

func TestIndexCARSignatureBlocks_RebuildsSignatureLinksFromCARBlocks(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, rootBytes := buildTestBlock(t)
	rootRawBlock, err := blocks.NewBlockWithCid(rootBytes, rootCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, rootRawBlock))

	sigCID, sigBytes := buildSignatureBlockForBlock(t, rootBlock)
	sigRawBlock, err := blocks.NewBlockWithCid(sigBytes, sigCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, sigRawBlock))

	signature, err := coreblock.GetSignatureBlockFromBytes(sigBytes)
	require.NoError(t, err)

	err = p.indexCARSignatureBlocks(ctx, []blocks.Block{rootRawBlock}, []blocks.Block{sigRawBlock})
	require.NoError(t, err)

	sigLink, found, err := coreblock.GetSignatureLink(
		ctx,
		multistore.Systemstore(),
		rootCID,
		string(signature.Header.Identity),
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sigCID, sigLink.Cid)
	require.NoError(t, p.verifyCARBlockSignatures(ctx, []blocks.Block{rootRawBlock}))
}

func TestIndexCARSignatureBlocks_IfSignatureDoesNotMatchRegularBlock_ShouldError(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, rootBytes := buildTestBlock(t)
	rootRawBlock, err := blocks.NewBlockWithCid(rootBytes, rootCID)
	require.NoError(t, err)

	otherBlock := rootBlock.Clone()
	otherBlock.Delta.DocCompositeDelta.DocID = []byte("otherDoc")
	sigCID, sigBytes := buildSignatureBlockForBlock(t, otherBlock)
	sigRawBlock, err := blocks.NewBlockWithCid(sigBytes, sigCID)
	require.NoError(t, err)

	err = p.indexCARSignatureBlocks(ctx, []blocks.Block{rootRawBlock}, []blocks.Block{sigRawBlock})
	require.Error(t, err)

	signatures, err := coreblock.ListSignatureLinks(ctx, multistore.Systemstore(), rootCID)
	require.NoError(t, err)
	require.Empty(t, signatures)
}

func TestStoreSignatureRecords_VerifiesBeforeIndexing(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, rootBytes := buildTestBlock(t)
	rootRawBlock, err := blocks.NewBlockWithCid(rootBytes, rootCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, rootRawBlock))

	sigCID, sigBytes := buildSignatureBlockForBlock(t, rootBlock)
	signature, err := coreblock.GetSignatureBlockFromBytes(sigBytes)
	require.NoError(t, err)

	err = p.storeSignatureRecords(ctx, []protocol.SignatureRecord{{
		BlockCID:     rootCID.Bytes(),
		SignatureCID: sigCID.Bytes(),
		Signature:    sigBytes,
	}})
	require.NoError(t, err)

	sigLink, found, err := coreblock.GetSignatureLink(
		ctx,
		multistore.Systemstore(),
		rootCID,
		string(signature.Header.Identity),
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sigCID, sigLink.Cid)
}

func TestStoreSignatureRecords_IfSignatureDoesNotMatchBlock_ShouldErrorAndNotIndex(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, rootBytes := buildTestBlock(t)
	rootRawBlock, err := blocks.NewBlockWithCid(rootBytes, rootCID)
	require.NoError(t, err)
	require.NoError(t, multistore.Blockstore().Put(ctx, rootRawBlock))

	otherBlock := rootBlock.Clone()
	otherBlock.Delta.DocCompositeDelta.DocID = []byte("otherDoc")
	sigCID, sigBytes := buildSignatureBlockForBlock(t, otherBlock)

	err = p.storeSignatureRecords(ctx, []protocol.SignatureRecord{{
		BlockCID:     rootCID.Bytes(),
		SignatureCID: sigCID.Bytes(),
		Signature:    sigBytes,
	}})
	require.Error(t, err)

	signatures, err := coreblock.ListSignatureLinks(ctx, multistore.Systemstore(), rootCID)
	require.NoError(t, err)
	require.Empty(t, signatures)
}

func TestStoreSignatureRecords_IfTargetBlockMissing_ShouldStoreSignatureBlockOnly(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	multistore := datastore.NewMultistore(rootstore, lockSet, immutable.None[int]())
	p := &P2P{
		db: &carTestDB{
			rootstore:  rootstore,
			multistore: multistore,
		},
	}

	rootBlock, rootCID, _ := buildTestBlock(t)
	sigCID, sigBytes := buildSignatureBlockForBlock(t, rootBlock)

	err := p.storeSignatureRecords(ctx, []protocol.SignatureRecord{{
		BlockCID:     rootCID.Bytes(),
		SignatureCID: sigCID.Bytes(),
		Signature:    sigBytes,
	}})
	require.NoError(t, err)

	_, err = multistore.Blockstore().Get(ctx, sigCID)
	require.NoError(t, err)

	signatures, err := coreblock.ListSignatureLinks(ctx, multistore.Systemstore(), rootCID)
	require.NoError(t, err)
	require.Empty(t, signatures)
}

func TestParseCAR_EmptyData(t *testing.T) {
	_, err := parseCAR([]byte{})
	require.Error(t, err)
}

func TestParseCAR_NilData(t *testing.T) {
	_, err := parseCAR(nil)
	require.Error(t, err)
}
