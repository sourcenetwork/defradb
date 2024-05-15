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

	dagsyncer "github.com/ipfs/boxo/fetcher"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/base"
	merklecrdt "github.com/sourcenetwork/defradb/internal/merkle/crdt"
)

type blockProcessor struct {
	*Peer
	txn       datastore.Txn
	col       client.Collection
	dsKey     core.DataStoreKey
	dagSyncer dagsyncer.Fetcher
	// List of composite blocks to eventually merge
	composites *list.List
}

func newBlockProcessor(
	p *Peer,
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	dagSyncer dagsyncer.Fetcher,
) *blockProcessor {
	return &blockProcessor{
		Peer:       p,
		composites: list.New(),
		txn:        txn,
		col:        col,
		dsKey:      dsKey,
		dagSyncer:  dagSyncer,
	}
}

// mergeBlock runs trough the list of composite blocks and sends them for processing.
func (bp *blockProcessor) mergeBlocks(ctx context.Context) {
	for e := bp.composites.Front(); e != nil; e = e.Next() {
		block := e.Value.(*coreblock.Block)
		link, _ := block.GenerateLink()
		err := bp.processBlock(ctx, block, link, "")
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to process block",
				err,
				corelog.String("DocID", bp.dsKey.DocID),
				corelog.Any("CID", link.Cid),
			)
		}
	}
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (bp *blockProcessor) processBlock(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
	field string,
) error {
	crdt, err := initCRDTForType(bp.txn, bp.col, bp.dsKey, field)
	if err != nil {
		return err
	}

	err = crdt.Clock().ProcessBlock(ctx, block, blockLink)
	if err != nil {
		return err
	}

	for _, link := range block.Links {
		if link.Name == core.HEAD {
			continue
		}

		b, err := bp.txn.DAGstore().Get(ctx, link.Cid)
		if err != nil {
			return err
		}

		childBlock, err := coreblock.GetFromBytes(b.RawData())
		if err != nil {
			return err
		}

		if err := bp.processBlock(ctx, childBlock, link.Link, link.Name); err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to process block",
				err,
				corelog.String("DocID", bp.dsKey.DocID),
				corelog.Any("CID", link.Cid),
			)
		}
	}

	return nil
}

func initCRDTForType(
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	field string,
) (merklecrdt.MerkleCRDT, error) {
	var key core.DataStoreKey
	var ctype client.CType
	description := col.Description()
	if field == "" { // empty field name implies composite type
		key = base.MakeDataStoreKeyWithCollectionDescription(
			description,
		).WithInstanceInfo(
			dsKey,
		).WithFieldId(
			core.COMPOSITE_NAMESPACE,
		)

		return merklecrdt.NewMerkleCompositeDAG(
			txn,
			core.NewCollectionSchemaVersionKey(col.Schema().VersionID, col.ID()),
			key,
			field,
		), nil
	}

	fd, ok := col.Definition().GetFieldByName(field)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find field %s for doc %s", field, dsKey.ToString()))
	}
	ctype = fd.Typ
	fieldID := fd.ID.String()
	key = base.MakeDataStoreKeyWithCollectionDescription(description).WithInstanceInfo(dsKey).WithFieldId(fieldID)

	return merklecrdt.InstanceWithStore(
		txn,
		core.NewCollectionSchemaVersionKey(col.Schema().VersionID, col.ID()),
		ctype,
		fd.Kind,
		key,
		field,
	)
}

// processRemoteBlock stores the block in the DAG store and initiates a sync of the block's children.
func (bp *blockProcessor) processRemoteBlock(
	ctx context.Context,
	session *sync.WaitGroup,
	block *coreblock.Block,
) error {
	link, err := block.GenerateLink()
	if err != nil {
		return err
	}

	b, err := block.Marshal()
	if err != nil {
		return err
	}

	if err := bp.txn.DAGstore().AsIPLDStorage().Put(ctx, link.Binary(), b); err != nil {
		return err
	}

	bp.handleChildBlocks(ctx, session, block)

	return nil
}

func (bp *blockProcessor) handleChildBlocks(
	ctx context.Context,
	session *sync.WaitGroup,
	block *coreblock.Block,
) {
	if block.Delta.IsComposite() {
		bp.composites.PushFront(block)
	}

	if len(block.Links) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, DAGSyncTimeout)
	defer cancel()

	for _, link := range block.Links {
		if !bp.queuedChildren.Visit(link.Cid) { // reserve for processing
			continue
		}

		exist, err := bp.txn.DAGstore().Has(ctx, link.Cid)
		if err != nil {
			log.ErrorContext(
				ctx,
				"Failed to check for existing block",
				corelog.Any("CID", link.Cid),
				corelog.Any("ERROR", err),
			)
		}
		if exist {
			continue
		}

		session.Add(1)
		job := &dagJob{
			session: session,
			cid:     link.Cid,
			bp:      bp,
		}

		select {
		case bp.sendJobs <- job:
		case <-bp.ctx.Done():
			return // jump out
		}
	}
}
