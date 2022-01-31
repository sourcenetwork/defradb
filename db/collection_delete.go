// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package db

import (
	"context"
	"errors"
	"fmt"

	block "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	dag "github.com/ipfs/go-merkledag"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

var (
	ErrInvalidDeleteTarget = errors.New("The doc delete targeter is an unknown type")
	ErrDeleteTargetEmpty   = errors.New("The doc delete targeter cannot be empty")
	ErrDeleteEmpty         = errors.New("The doc delete cannot be empty")
)

// DeleteWith deletes a target document using the given deleter type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Delete.
// Eg: DeleteWithFilter or DeleteWithKey
func (c *Collection) DeleteWith(ctx context.Context, target interface{}, opts ...client.DeleteOpt) error {
	// switch t := target.(type) {
	// case string, map[string]interface{}, *parser.Filter:
	// _, err := c.DeleteWithFilter(ctx, t, deleter, opts...)
	// return err
	// case key.DocKey:
	// _, err := c.DeleteWithKey(ctx, t, deleter, opts...)
	// return err
	// case []key.DocKey:
	// _, err := c.DeleteWithKeys(ctx, t, deleter, opts...)
	// return err
	// case *document.SimpleDocument:
	// return c.DeleteWithDoc(t, deleter, opts...)
	// case []*document.SimpleDocument:
	// return c.DeleteWithDocs(t, deleter, opts...)
	// default:
	// return ErrInvalidTarget
	// }
	return nil
}

// DeleteWithFilter deletes using a filter to target documents for delete.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithFilter(ctx context.Context, filter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// txn, err := c.getTxn(ctx, false)
	// if err != nil {
	// 	return nil, err
	// }
	// defer c.discardImplicitTxn(ctx, txn)
	// res, err := c.deleteWithFilter(ctx, txn, filter, deleter, opts...)
	// if err != nil {
	// 	return nil, err
	// }
	// return res, c.commitImplicitTxn(ctx, txn)

	return nil, nil
}

// DeleteWithKey deletes using a DocKey to target a single document for delete.
func (c *Collection) DeleteWithKey(ctx context.Context, key key.DocKey, opts ...client.DeleteOpt) (*client.DeleteResult, error) {

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	defer c.discardImplicitTxn(ctx, txn)

	res, err := c.deleteWithKey(ctx, txn, key, opts...)

	fmt.Println("--------------------------------------")
	fmt.Println("res: ", res)
	fmt.Println("err: ", err)
	fmt.Println("--------------------------------------")

	// res, err := c.deleteWithKey(ctx, txn, key, deleter, opts...)
	// if err != nil {
	// 	return nil, err
	// }
	// return res, c.commitImplicitTxn(ctx, txn)

	return nil, nil
}

func (c *Collection) deleteWithKey(ctx context.Context, txn *Txn, key key.DocKey, opts ...client.DeleteOpt) (*client.DeleteResult, error) {

	doc, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	v, err := doc.ToMap()
	if err != nil {
		return nil, err
	}

	err = c.applyDelete(ctx, txn, v)

	results := &client.DeleteResult{
		Count:   1,
		DocKeys: []string{key.String()},
	}
	return results, nil

}

func (c *Collection) applyDelete(ctx context.Context, txn *Txn, doc map[string]interface{}) error {
	// keyStr, ok := doc["_key"].(string)
	// if !ok {
	// 	return errors.New("Document is missing key")
	// }
	// key := ds.NewKey(keyStr)
	// links := make([]core.DAGLink, 0)
	// for mfield, mval := range merge {
	// 	if _, ok := mval.(map[string]interface{}); ok {
	// 		return ErrInvalidMergeValueType
	// 	}

	// 	fd, valid := c.desc.GetField(mfield)
	// 	if !valid {
	// 		return errors.New("Invalid field in Patch")
	// 	}

	// 	cval, err := validateFieldSchema(mval, fd)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	val := document.NewCBORValue(fd.Typ, cval)
	// 	fieldKey := c.getFieldKey(key, mfield)
	// 	c, err := c.saveDocValue(ctx, txn, c.getPrimaryIndexDocKey(fieldKey), val)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// links[mfield] = c
	// 	links = append(links, core.DAGLink{
	// 		Name: mfield,
	// 		Cid:  c,
	// 	})
	// }

	// // Update CompositeDAG
	// em, err := cbor.CanonicalEncOptions().EncMode()
	// if err != nil {
	// 	return err
	// }
	// buf, err := em.Marshal(merge)
	// if err != nil {
	// 	return err
	// }
	// if _, err := c.saveValueToMerkleCRDT(ctx, txn, c.getPrimaryIndexDocKey(key), core.COMPOSITE, buf, links); err != nil {
	// 	return err
	// }

	// // if this a a Batch masked as a Transaction
	// // commit our writes so we can see them.
	// // Batches don't maintain serializability, or
	// // linearization, or any other transaction
	// // semantics, which the user already knows
	// // otherwise they wouldn't use a datastore
	// // that doesn't support proper transactions.
	// // So lets just commit, and keep going.
	// // @todo: Change this on the Txn.BatchShim
	// // structure
	// if txn.IsBatch() {
	// 	if err := txn.Commit(ctx); err != nil {
	// 		return err
	// 	}
	// }

	// return nil

	return nil
}

// =================================== UNIMPLEMENTED ===================================

// DeleteWithKeys is the same as DeleteWithKey but accepts multiple keys as a slice.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithKeys(ctx context.Context, keys []key.DocKey, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// txn, err := c.getTxn(ctx, false)
	// if err != nil {
	// return nil, err
	// }
	// defer c.discardImplicitTxn(ctx, txn)
	// res, err := c.deleteWithKeys(ctx, txn, keys, deleter, opts...)
	// if err != nil {
	// return nil, err
	// }
	// return res, c.commitImplicitTxn(ctx, txn)

	return nil, nil
}

// DeleteWithDoc deletes targeting the supplied document.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithDoc(doc *document.SimpleDocument, opts ...client.DeleteOpt) error {
	return nil
}

// DeleteWithDocs deletes all the supplied documents in the slice.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithDocs(docs []*document.SimpleDocument, opts ...client.DeleteOpt) error {
	return nil
}

func (c *Collection) deleteWithKeys(ctx context.Context, txn *Txn, keys []key.DocKey, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// fmt.Println("updating keys:", keys)
	// patch, err := parseDeleter(deleter)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// isPatch := false
	// switch patch.(type) {
	// case []map[string]interface{}:
	// 	isPatch = true
	// case map[string]interface{}:
	// 	isPatch = false
	// default:
	// 	return nil, ErrInvalidDeleter
	// }
	//
	// results := &client.DeleteResult{
	// 	DocKeys: make([]string, len(keys)),
	// }
	// for i, key := range keys {
	// 	doc, err := c.Get(ctx, key)
	// 	if err != nil {
	// 		fmt.Println("error getting key to delete:", key)
	// 		return nil, err
	// 	}
	// 	v, err := doc.ToMap()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if isPatch {
	// 		// todo
	// 	} else {
	// 		err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
	// 	}
	// 	if err != nil {
	// 		return nil, nil
	// 	}
	//
	// 	results.DocKeys[i] = key.String()
	// 	results.Count++
	// }
	// return results, nil

	return nil, nil
}

func (c *Collection) deleteWithFilter(ctx context.Context, txn *Txn, filter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// patch, err := parseDeleter(deleter)
	// if err != nil {
	// 	return nil, err
	// }

	// isPatch := false
	// isMerge := false
	// switch patch.(type) {
	// case []map[string]interface{}:
	// 	isPatch = true
	// case map[string]interface{}:
	// 	isMerge = true
	// default:
	// 	return nil, ErrInvalidDeleter
	// }

	// // scan through docs with filter
	// query, err := c.makeSelectionQuery(ctx, txn, filter, opts...)
	// if err != nil {
	// 	return nil, err
	// }
	// if err := query.Start(); err != nil {
	// 	return nil, err
	// }

	// results := &client.DeleteResult{
	// 	DocKeys: make([]string, 0),
	// }

	// // loop while we still have results from the filter query
	// for {
	// 	next, err := query.Next()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// if theres no more records from the query, jump out of the loop
	// 	if !next {
	// 		break
	// 	}

	// 	// Get the document, and apply the patch
	// 	doc := query.Values()
	// 	if isPatch {
	// 		err = c.applyPatch(txn, doc, patch.([]map[string]interface{}))
	// 	} else if isMerge { // else is fine here
	// 		err = c.applyMerge(ctx, txn, doc, patch.(map[string]interface{}))
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// add successful deleted doc to results
	// 	results.DocKeys = append(results.DocKeys, doc["_key"].(string))
	// 	results.Count++
	// }

	// return results, nil

	return nil, nil
}

func (c *Collection) deleteFull(ctx context.Context, txn *Txn, dockey key.DocKey) error {
	// quick check doc exists via c.exists(ctx, txn, dockey)

	// get head cid
	// dockey => bae-kljhLKHJG-lkjhgkldjhlzkdf-kdhflkhjsklgh-kjdhlkghjs
	// key => bae-kljhLKHJG-lkjhgkldjhlzkdf-kdhflkhjsklgh-kjdhlkghjs/C
	// /db
	// -> datastore /data => /db/data
	// -> systemstore /system => /db/system
	// -> block /blocks => /db/blocks
	// -> headstore /heads => /db/heads
	// var key ds.Key
	key := dockey.Key.ChildString(core.COMPOSITE_NAMESPACE)
	headset := clock.NewHeadSet(txn.Headstore(), key)
	heads, _, err := headset.List(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get document heads: %w", err)
	}

	// docs: https://pkg.go.dev/github.com/ipfs/go-datastore
	// 1. delete datastore state (txn.Datastore.Query({Prefix: c.GetPrimaryIndexDocKey(dockey)})) -> loop over results, and delete
	// 2. delete headstore state (txn.Headstore.Query({Prefix: dockey.Key})) - > loop over and delete

	// delete block state
	// /db/blocks/CIQSDFKLJGHFKLSJGHHJKKLGHGLHSKLHKJGS => KLJSFHGLKJFHJKDLGKHDGLHGLFDHGLFDGKGHL
	dDel := newDagDeleter(txn.DAGstore())
	// for head in heads => dDel.run(ctx, head)
	return dDel.run(ctx, heads[0])
}

//						   		   | --> (x) HEAD#1->cid1
// (xx) --> (xx) --> (xx) --> (xx) |
//						   		   | --> (x) HEAD#2->cid2

type dagDeleter struct {
	bstore core.DAGStore
	// queue *list.List
}

func newDagDeleter(bstore core.DAGStore) dagDeleter {
	return dagDeleter{
		bstore: bstore,
	}
}

func (d dagDeleter) run(ctx context.Context, c cid.Cid) error {
	// base case ?

	// nil check
	if c == cid.Undef {
		return nil
	}

	// get block
	blk, err := d.bstore.Get(ctx, c)
	if err == blockstore.ErrNotFound {
		return nil
	} else if err != nil {
		return err
	}

	// call delete
	return d.delete(ctx, c, blk)

}

//
//  (ipld.Block(ipldProtobufNode{Data: (cbor(crdt deltaPayload)), Links: (_head => parentCid, fieldName => fieldCid)))
//
func (d dagDeleter) delete(ctx context.Context, c cid.Cid, blk block.Block) error {
	nd, err := dag.DecodeProtobuf(blk.RawData())
	if err != nil {
		return err
	}

	// delete current block
	if err := d.bstore.DeleteBlock(ctx, c); err != nil {
		return err
	}

	for _, link := range nd.Links() {
		// link.Name, link.Cid
		if err := d.run(ctx, link.Cid); err != nil {
			return err
		}
	}

	return nil
}
