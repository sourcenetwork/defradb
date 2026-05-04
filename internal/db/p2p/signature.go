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
	"context"
	"sort"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

func (p *P2P) newPushLogRequest(
	ctx context.Context,
	docID string,
	blockCID cid.Cid,
	collectionID string,
	blockBytes []byte,
) (protocol.PushLogRequest, error) {
	signatures, err := p.collectSignatureRecords(ctx, blockCID, blockBytes)
	if err != nil {
		return protocol.PushLogRequest{}, err
	}

	return protocol.PushLogRequest{
		CollectionID: collectionID,
		Creator:      p.host.ID(),
		Documents: []protocol.DocumentInfo{{
			DocID: docID,
			CID:   blockCID.Bytes(),
			Block: blockBytes,
		}},
		Signatures: signatures,
	}, nil
}

func (p *P2P) collectSignatureRecordsForDocuments(
	ctx context.Context,
	documents []protocol.DocumentInfo,
) ([]protocol.SignatureRecord, error) {
	recordsBySignatureCID := map[string]protocol.SignatureRecord{}

	for _, doc := range documents {
		blockCID, err := cid.Cast(doc.CID)
		if err != nil {
			return nil, err
		}

		records, err := p.collectSignatureRecords(ctx, blockCID, doc.Block)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			recordsBySignatureCID[string(record.SignatureCID)] = record
		}
	}

	records := make([]protocol.SignatureRecord, 0, len(recordsBySignatureCID))
	for _, record := range recordsBySignatureCID {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return string(records[i].SignatureCID) < string(records[j].SignatureCID)
	})

	return records, nil
}

func (p *P2P) collectSignatureRecords(
	ctx context.Context,
	blockCID cid.Cid,
	blockBytes []byte,
) ([]protocol.SignatureRecord, error) {
	visited := map[string]struct{}{}
	recordsBySignatureCID := map[string]protocol.SignatureRecord{}

	var collectForBlock func(cid.Cid, []byte) error
	collectForBlock = func(currentCID cid.Cid, currentBlockBytes []byte) error {
		if _, ok := visited[currentCID.String()]; ok {
			return nil
		}
		visited[currentCID.String()] = struct{}{}

		signatures, err := coreblock.ListSignatureLinks(ctx, p.db.Multistore().Systemstore(), currentCID)
		if err != nil {
			return err
		}
		for _, signature := range signatures {
			if _, ok := recordsBySignatureCID[signature.Link.Cid.String()]; ok {
				continue
			}

			rawBlock, err := p.db.Multistore().Blockstore().Get(ctx, signature.Link.Cid)
			if err != nil {
				return err
			}

			recordsBySignatureCID[signature.Link.Cid.String()] = protocol.SignatureRecord{
				BlockCID:     currentCID.Bytes(),
				SignatureCID: signature.Link.Cid.Bytes(),
				Signature:    rawBlock.RawData(),
			}
		}

		block, err := coreblock.GetFromBytes(currentBlockBytes)
		if err != nil {
			return err
		}

		for _, link := range block.AllLinks() {
			rawBlock, err := p.db.Multistore().Blockstore().Get(ctx, link.Cid)
			if err != nil {
				return err
			}
			if err := collectForBlock(link.Cid, rawBlock.RawData()); err != nil {
				return err
			}
		}

		return nil
	}

	if err := collectForBlock(blockCID, blockBytes); err != nil {
		return nil, err
	}

	records := make([]protocol.SignatureRecord, 0, len(recordsBySignatureCID))
	for _, record := range recordsBySignatureCID {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return string(records[i].SignatureCID) < string(records[j].SignatureCID)
	})

	return records, nil
}

func (p *P2P) storeSignatureRecords(ctx context.Context, records []protocol.SignatureRecord) error {
	for _, record := range records {
		blockCID, err := cid.Cast(record.BlockCID)
		if err != nil {
			return err
		}

		signatureCID, err := cid.Cast(record.SignatureCID)
		if err != nil {
			return err
		}

		computedCID, err := signatureCID.Prefix().Sum(record.Signature)
		if err != nil {
			return err
		}
		if !computedCID.Equals(signatureCID) {
			return errors.New("signature CID does not match signature bytes")
		}

		signature, err := coreblock.GetSignatureBlockFromBytes(record.Signature)
		if err != nil {
			return err
		}

		rawBlock, err := blocks.NewBlockWithCid(record.Signature, signatureCID)
		if err != nil {
			return err
		}
		if err := p.db.Multistore().Blockstore().Put(ctx, rawBlock); err != nil {
			return err
		}

		err = coreblock.StoreSignatureLink(
			ctx,
			p.db.Multistore().Systemstore(),
			blockCID,
			signature,
			cidlink.Link{Cid: signatureCID},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
