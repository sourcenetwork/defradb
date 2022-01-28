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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

var (
	ErrInvalidDeleteTarget = errors.New("The doc delete targeter is an unknown type")
	ErrDeleteTargetEmpty   = errors.New("The doc delete targeter cannot be empty")
	ErrInvalidDeleter      = errors.New("The doc deleter is an unknown type")
	ErrDeleteEmpty         = errors.New("The doc delete cannot be empty")
)

// Delete2 deletes the given doc. It will scan through the field/value pairs
// and find those marked for delete, and apply the appropriate delete.
// Delete only works on root level field/value pairs. So not foreign or related
// types can be deleted. If you wish to delete sub types, use DeleteWith, and supply
// an delete payload in the form of a Patch or a Merge Patch.
func (c *Collection) Delete2(doc *document.SimpleDocument, opts ...client.DeleteOpt) error {
	return nil
}

// DeleteWith deletes a target document using the given deleter type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Delete.
// Eg: DeleteWithFilter or DeleteWithKey
func (c *Collection) DeleteWith(ctx context.Context, target interface{}, deleter interface{}, opts ...client.DeleteOpt) error {
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
func (c *Collection) DeleteWithFilter(ctx context.Context, filter interface{}, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
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
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithKey(ctx context.Context, key key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// txn, err := c.getTxn(ctx, false)
	// if err != nil {
	// return nil, err
	// }
	// defer c.discardImplicitTxn(ctx, txn)
	// res, err := c.deleteWithKey(ctx, txn, key, deleter, opts...)
	// if err != nil {
	// return nil, err
	// }
	// return res, c.commitImplicitTxn(ctx, txn)

	return nil, nil
}

// DeleteWithKeys is the same as DeleteWithKey but accepts multiple keys as a slice.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithKeys(ctx context.Context, keys []key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
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
func (c *Collection) DeleteWithDoc(doc *document.SimpleDocument, deleter interface{}, opts ...client.DeleteOpt) error {
	return nil
}

// DeleteWithDocs deletes all the supplied documents in the slice.
// An deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithDocs(docs []*document.SimpleDocument, deleter interface{}, opts ...client.DeleteOpt) error {
	return nil
}

func (c *Collection) deleteWithKey(ctx context.Context, txn *Txn, key key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	// patch, err := parseDeleter(deleter)
	// if err != nil {
	// 	return nil, err
	// }

	// isPatch := false
	// switch patch.(type) {
	// case []map[string]interface{}:
	// 	isPatch = true
	// case map[string]interface{}:
	// 	isPatch = false
	// default:
	// 	return nil, ErrInvalidDeleter
	// }

	// doc, err := c.Get(ctx, key)
	// if err != nil {
	// 	return nil, err
	// }
	// v, err := doc.ToMap()
	// if err != nil {
	// 	return nil, err
	// }

	// if isPatch {
	// 	// todo
	// } else {
	// 	err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
	// }
	// if err != nil {
	// 	return nil, err
	// }

	// results := &client.DeleteResult{
	// 	Count:   1,
	// 	DocKeys: []string{key.String()},
	// }
	// return results, nil

	return nil, nil
}

func (c *Collection) deleteWithKeys(ctx context.Context, txn *Txn, keys []key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
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

func (c *Collection) deleteWithFilter(ctx context.Context, txn *Txn, filter interface{}, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
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
