// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	merklecrdt "github.com/sourcenetwork/defradb/merkle/crdt"
)

type blockProcessor struct {
	*Peer
	txn    datastore.Txn
	col    client.Collection
	dsKey  core.DataStoreKey
	getter ipld.NodeGetter
	// List of composite blocks to eventually merge
	composites *list.List
}

func newBlockProcessor(
	p *Peer,
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	getter ipld.NodeGetter,
) *blockProcessor {
	return &blockProcessor{
		Peer:       p,
		composites: list.New(),
		txn:        txn,
		col:        col,
		dsKey:      dsKey,
		getter:     getter,
	}
}

// mergeBlock runs trough the list of composite blocks and sends them for processing.
func (bp *blockProcessor) mergeBlocks(ctx context.Context) {
	for e := bp.composites.Front(); e != nil; e = e.Next() {
		nd := e.Value.(ipld.Node)
		err := bp.processBlock(ctx, nd, "")
		if err != nil {
			log.ErrorE(
				ctx,
				"Failed to process block",
				err,
				logging.NewKV("DocID", bp.dsKey.DocID),
				logging.NewKV("CID", nd.Cid()),
			)
		}
	}
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (bp *blockProcessor) processBlock(ctx context.Context, nd ipld.Node, field string) error {
	crdt, err := initCRDTForType(ctx, bp.txn, bp.col, bp.dsKey, field)
	if err != nil {
		return err
	}
	delta, err := crdt.DeltaDecode(nd)
	if err != nil {
		return errors.Wrap("failed to decode delta object", err)
	}

	err = crdt.Clock().ProcessNode(ctx, delta, nd)
	if err != nil {
		return err
	}

	for _, link := range nd.Links() {
		if link.Name == core.HEAD {
			continue
		}

		block, err := bp.txn.DAGstore().Get(ctx, link.Cid)
		if err != nil {
			return err
		}
		nd, err := dag.DecodeProtobufBlock(block)
		if err != nil {
			return err
		}

		if err := bp.processBlock(ctx, nd, link.Name); err != nil {
			log.ErrorE(
				ctx,
				"Failed to process block",
				err,
				logging.NewKV("DocID", bp.dsKey.DocID),
				logging.NewKV("CID", nd.Cid()),
			)
		}
	}

	return nil
}

func initCRDTForType(
	ctx context.Context,
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	field string,
) (merklecrdt.MerkleCRDT, error) {
	var key core.DataStoreKey
	var ctype client.CType
	description := col.Description()
	if field == "" { // empty field name implies composite type
		ctype = client.COMPOSITE
		key = base.MakeDSKeyWithCollectionID(
			description,
		).WithInstanceInfo(
			dsKey,
		).WithFieldId(
			core.COMPOSITE_NAMESPACE,
		)

		log.Debug(ctx, "Got CRDT Type", logging.NewKV("CType", ctype), logging.NewKV("Field", field))
		return merklecrdt.NewMerkleCompositeDAG(
			txn,
			core.NewCollectionSchemaVersionKey(col.Schema().VersionID, col.ID()),
			key,
			field,
		), nil
	}

	fd, ok := col.Schema().GetField(field)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find field %s for doc %s", field, dsKey))
	}
	ctype = fd.Typ
	fieldID := fd.ID.String()
	key = base.MakeDSKeyWithCollectionID(description).WithInstanceInfo(dsKey).WithFieldId(fieldID)

	log.Debug(ctx, "Got CRDT Type", logging.NewKV("CType", ctype), logging.NewKV("Field", field))
	return merklecrdt.NewMerkleLWWRegister(
		txn,
		core.NewCollectionSchemaVersionKey(col.Schema().VersionID, col.ID()),
		key,
		field,
	), nil
}

func decodeBlockBuffer(buf []byte, cid cid.Cid) (ipld.Node, error) {
	blk, err := blocks.NewBlockWithCid(buf, cid)
	if err != nil {
		return nil, errors.Wrap("failed to create block", err)
	}
	return ipld.Decode(blk, dag.DecodeProtobufBlock)
}

// processRemoteBlock stores the block in the DAG store and initiates a sync of the block's children.
func (bp *blockProcessor) processRemoteBlock(
	ctx context.Context,
	session *sync.WaitGroup,
	nd ipld.Node,
	isComposite bool,
) error {
	log.Debug(ctx, "Running processLog")

	if err := bp.txn.DAGstore().Put(ctx, nd); err != nil {
		return err
	}

	if isComposite {
		bp.composites.PushFront(nd)
	}

	bp.handleChildBlocks(ctx, session, nd, isComposite)

	return nil
}

func (bp *blockProcessor) handleChildBlocks(
	ctx context.Context,
	session *sync.WaitGroup,
	nd ipld.Node,
	isComposite bool,
) {
	if len(nd.Links()) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, DAGSyncTimeout)
	defer cancel()

	for _, link := range nd.Links() {
		if !bp.queuedChildren.Visit(link.Cid) { // reserve for processing
			continue
		}

		exist, err := bp.txn.DAGstore().Has(ctx, link.Cid)
		if err != nil {
			log.Error(
				ctx,
				"Failed to check for existing block",
				logging.NewKV("CID", link.Cid),
				logging.NewKV("ERROR", err),
			)
		}
		if exist {
			log.Debug(ctx, "Already have block locally, skipping.", logging.NewKV("CID", link.Cid))
			continue
		}

		session.Add(1)
		job := &dagJob{
			session:     session,
			cid:         link.Cid,
			isComposite: isComposite && link.Name == core.HEAD,
			bp:          bp,
		}

		select {
		case bp.sendJobs <- job:
		case <-bp.ctx.Done():
			return // jump out
		}
	}
}
