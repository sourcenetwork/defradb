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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// processNode is a general utility for processing various kinds
// of CRDT blocks
func (p *Peer) processLog(
	ctx context.Context,
	col client.Collection,
	dockey core.Key,
	c cid.Cid,
	field string,
	nd ipld.Node,
	getter format.NodeGetter) ([]cid.Cid, error) {
	log.Debug(ctx, "Running processLog")

	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	// KEEPING FOR REFERENCE FOR NOW
	// check if we already have this block
	// exists, err := txn.DAGstore().Has(ctx, c)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to check for existing block %s: %w", c, err)
	// }
	// if exists {
	// 	log.Debugf("Already have block %s locally, skipping.", c)
	// 	return nil, nil
	// }

	crdt, err := initCRDTForType(ctx, txn, col, dockey, field)
	if err != nil {
		return nil, err
	}

	delta, err := crdt.DeltaDecode(nd)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode delta object: %w", err)
	}

	log.Debug(ctx, "Processing push log request", logging.NewKV("DocKey", dockey), logging.NewKV("Cid", c))
	height := delta.GetPriority()

	if err := txn.DAGstore().Put(ctx, nd); err != nil {
		return nil, err
	}

	ng := p.createNodeGetter(crdt, getter)
	cids, err := crdt.Clock().ProcessNode(ctx, ng, c, height, delta, nd)
	if err != nil {
		return nil, err
	}

	// mark this obj as done
	p.queuedChildren.Remove(c)

	return cids, txn.Commit(ctx)
}

func initCRDTForType(ctx context.Context, txn core.MultiStore, col client.Collection, docKey core.Key, field string) (crdt.MerkleCRDT, error) {
	var key core.Key
	var ctype core.CType
	if field == "" { // empty field name implies composite type
		ctype = core.COMPOSITE
		key = core.NewKey(col.GetPrimaryIndexDocKey(docKey).ChildString(core.COMPOSITE_NAMESPACE).String())
	} else {
		fd, ok := col.Description().GetField(field)
		if !ok {
			return nil, fmt.Errorf("Couldn't find field %s for doc %s", field, docKey)
		}
		ctype = fd.Typ
		fieldID := fd.ID.String()
		key = core.NewKey(col.GetPrimaryIndexDocKey(docKey).ChildString(fieldID).String())
	}
	log.Debug(ctx, "Got CRDT Type", logging.NewKV("CType", ctype), logging.NewKV("Field", field))
	return crdt.DefaultFactory.InstanceWithStores(txn, col.SchemaID(), nil, ctype, key)
}

func decodeBlockBuffer(buf []byte, cid cid.Cid) (ipld.Node, error) {
	blk, err := blocks.NewBlockWithCid(buf, cid)
	if err != nil {
		return nil, fmt.Errorf("Failed to create block: %w", err)
	}
	return format.Decode(blk)
}

func (p *Peer) createNodeGetter(crdt crdt.MerkleCRDT, getter format.NodeGetter) *clock.CrdtNodeGetter {
	return &clock.CrdtNodeGetter{
		NodeGetter:     getter,
		DeltaExtractor: crdt.DeltaDecode,
	}
}

func (p *Peer) handleChildBlocks(
	session *sync.WaitGroup,
	col client.Collection,
	dockey core.Key,
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
			log.ErrorE(ctx, "Failed to get node", err, logging.NewKV("Cid", c))
			continue
		}

		log.Debug(
			ctx,
			"Submitting new job to dag queue",
			logging.NewKV("Collection", col.Name()),
			logging.NewKV("DocKey", dockey),
			logging.NewKV("Field", fieldName),
			logging.NewKV("Cid", cNode.Cid()))

		session.Add(1)
		job := &dagJob{
			collection: col,
			dockey:     dockey,
			fieldName:  fieldName,
			session:    session,
			nodeGetter: getter,
			node:       cNode,
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
