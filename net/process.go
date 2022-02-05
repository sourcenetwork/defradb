package net

import (
	"context"
	"fmt"
	"sync"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

// processNode is a general utility for processing various kinds
// of CRDT blocks
func (p *Peer) processLog(
	ctx context.Context,
	col client.Collection,
	dockey key.DocKey,
	c cid.Cid,
	field string,
	nd ipld.Node,
	getter format.NodeGetter) ([]cid.Cid, error) {
	log.Debugf("running processLog")
	// check if we already have this block
	// exists, err := txn.DAGstore().Has(ctx, c)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to check for existing block %s: %w", c, err)
	// }
	// if exists {
	// 	log.Debugf("Already have block %s locally, skipping.", c)
	// 	return nil, nil
	// }

	txn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	crdt, err := initCRDTForType(txn, col, dockey, field)
	if err != nil {
		return nil, err
	}

	delta, err := crdt.DeltaDecode(nd)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode delta object: %w", err)
	}

	log.Debugf("Processing push log request for %s at %s", dockey, c)
	height := delta.GetPriority()

	if err := txn.DAGstore().Put(ctx, nd); err != nil {
		return nil, err
	}

	ng := p.createNodeGetter(crdt, getter)
	cid, err := crdt.Clock().ProcessNode(ctx, ng, c, height, delta, nd)
	if err != nil {
		return nil, err
	}
	return cid, txn.Commit(ctx)
}

func initCRDTForType(txn core.MultiStore, col client.Collection, docKey key.DocKey, field string) (crdt.MerkleCRDT, error) {
	var key ds.Key
	var ctype core.CType
	if field == "" { // empty field name implies composite type
		ctype = core.COMPOSITE
		key = col.GetPrimaryIndexDocKey(docKey.Key).ChildString(core.COMPOSITE_NAMESPACE)
	} else {
		fd, ok := col.Description().GetField(field)
		if !ok {
			return nil, fmt.Errorf("Couldn't find field %s for doc %s", field, docKey)
		}
		ctype = fd.Typ
		fieldID := fd.ID.String()
		key = col.GetPrimaryIndexDocKey(docKey.Key).ChildString(fieldID)
	}
	log.Debugf("Got CRDT Type: %v for %s", ctype, field)
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

// func (p *Peer) processComposite

func (p *Peer) handleChildBlocks(
	session *sync.WaitGroup,
	col client.Collection,
	dockey key.DocKey,
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
		var fieldName string
		// loop over our children to get the cooresponding field names from the DAG
		for _, l := range nd.Links() {
			if c == l.Cid {
				if l.Name != core.HEAD {
					fieldName = l.Name
				}
			}
		}
		// @todo: handle no match case ^^

		// heads of subfields are still subfields, not composites
		if fieldName == "" && field != "" {
			fieldName = field
		}

		// get object
		cNode, err := getter.Get(ctx, c)
		if err != nil {
			log.Errorf("Failed to get node %s: %s", c, err)
			continue
		}

		log.Debugf("Submitting new job to dag queue - col: %s, key: %s, field: %s, cid: %s", col.Name(), dockey, fieldName, cNode.Cid())
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
			session.Done()
			return
		}

	}

	// Clear up any children we failed to get from queued children
	// for _, child := range children {
	// 	p.queuedChildren.Remove(child)
	// }
}
