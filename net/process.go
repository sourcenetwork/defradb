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
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// processNode is a general utility for processing various kinds
// of CRDT blocks
func (p *Peer) processLog(
	ctx context.Context,
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	field string,
	nd ipld.Node,
	getter ipld.NodeGetter,
	removeChildren bool,
) ([]cid.Cid, error) {
	log.Debug(ctx, "Running processLog")

	crdt, err := initCRDTForType(ctx, txn, col, dsKey, field)
	if err != nil {
		return nil, err
	}

	delta, err := crdt.DeltaDecode(nd)
	if err != nil {
		return nil, errors.Wrap("failed to decode delta object", err)
	}

	log.Debug(
		ctx,
		"Processing PushLog request",
		logging.NewKV("Datastore key", dsKey),
		logging.NewKV("CID", nd.Cid()),
	)

	if err := txn.DAGstore().Put(ctx, nd); err != nil {
		return nil, err
	}

	ng := p.createNodeGetter(crdt, getter)
	cids, err := crdt.Clock().ProcessNode(ctx, ng, delta, nd)
	if err != nil {
		return nil, err
	}

	if removeChildren {
		// mark this obj as done
		p.queuedChildren.Remove(nd.Cid())
	}

	return cids, nil
}

func initCRDTForType(
	ctx context.Context,
	txn datastore.MultiStore,
	col client.Collection,
	dsKey core.DataStoreKey,
	field string,
) (crdt.MerkleCRDT, error) {
	var key core.DataStoreKey
	var ctype client.CType
	description := col.Description()
	if field == "" { // empty field name implies composite type
		ctype = client.COMPOSITE
		key = base.MakeCollectionKey(
			description,
		).WithInstanceInfo(
			dsKey,
		).WithFieldId(
			core.COMPOSITE_NAMESPACE,
		)
	} else {
		fd, ok := col.Schema().GetField(field)
		if !ok {
			return nil, errors.New(fmt.Sprintf("Couldn't find field %s for doc %s", field, dsKey))
		}
		ctype = fd.Typ
		fieldID := fd.ID.String()
		key = base.MakeCollectionKey(description).WithInstanceInfo(dsKey).WithFieldId(fieldID)
	}
	log.Debug(ctx, "Got CRDT Type", logging.NewKV("CType", ctype), logging.NewKV("Field", field))
	return crdt.DefaultFactory.InstanceWithStores(
		txn,
		core.NewCollectionSchemaVersionKey(col.Schema().VersionID, col.ID()),
		events.EmptyUpdateChannel,
		ctype,
		key,
		field,
	)
}

func decodeBlockBuffer(buf []byte, cid cid.Cid) (ipld.Node, error) {
	blk, err := blocks.NewBlockWithCid(buf, cid)
	if err != nil {
		return nil, errors.Wrap("failed to create block", err)
	}
	return ipld.Decode(blk, dag.DecodeProtobufBlock)
}

func (p *Peer) createNodeGetter(
	crdt crdt.MerkleCRDT,
	getter ipld.NodeGetter,
) *clock.CrdtNodeGetter {
	return &clock.CrdtNodeGetter{
		NodeGetter:     getter,
		DeltaExtractor: crdt.DeltaDecode,
	}
}

func (p *Peer) handleChildBlocks(
	session *sync.WaitGroup,
	txn datastore.Txn,
	col client.Collection,
	dsKey core.DataStoreKey,
	field string,
	nd ipld.Node,
	children []cid.Cid,
	getter ipld.NodeGetter,
) {
	if len(children) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(p.ctx, DAGSyncTimeout)
	defer cancel()

	for _, c := range children {
		if !p.queuedChildren.Visit(c) { // reserve for processing
			continue
		}

		var fieldName string
		// loop over our children to get the corresponding field names from the DAG
		for _, l := range nd.Links() {
			if c == l.Cid {
				if l.Name != core.HEAD {
					fieldName = l.Name
				}
			}
		}

		// heads of subfields are still subfields, not composites
		if fieldName == "" && field != "" {
			fieldName = field
		}

		// get object
		cNode, err := getter.Get(ctx, c)
		if err != nil {
			log.ErrorE(ctx, "Failed to get node", err, logging.NewKV("CID", c))
			continue
		}

		log.Debug(
			ctx,
			"Submitting new job to DAG queue",
			logging.NewKV("Collection", col.Name()),
			logging.NewKV("Datastore key", dsKey),
			logging.NewKV("Field", fieldName),
			logging.NewKV("CID", cNode.Cid()))

		session.Add(1)
		job := &dagJob{
			collection: col,
			dsKey:      dsKey,
			fieldName:  fieldName,
			session:    session,
			nodeGetter: getter,
			node:       cNode,
			txn:        txn,
		}

		select {
		case p.sendJobs <- job:
		case <-p.ctx.Done():
			return // jump out
		}
	}

	// Clear up any children we failed to get from queued children
	// for _, child := range children {
	// 	p.queuedChildren.Remove(child)
	// }
}
