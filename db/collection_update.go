// Copyright 2022 Democratized Data Foundation
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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/planner"

	cbor "github.com/fxamacker/cbor/v2"
)

var (
	ErrInvalidUpdateTarget   = errors.New("The doc update targeter is an unknown type")
	ErrUpdateTargetEmpty     = errors.New("The doc update targeter cannot be empty")
	ErrInvalidUpdater        = errors.New("The doc updater is an unknown type")
	ErrUpdateEmpty           = errors.New("The doc update cannot be empty")
	ErrInvalidMergeValueType = errors.New("The type of value in the merge patch doesn't match the schema")
)

// UpdateWith updates a target document using the given updater type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Update.
// Eg: UpdateWithFilter or UpdateWithKey
func (c *collection) UpdateWith(ctx context.Context, target interface{}, updater interface{}, opts ...client.UpdateOpt) error {
	switch t := target.(type) {
	case string, map[string]interface{}, *parser.Filter:
		_, err := c.UpdateWithFilter(ctx, t, updater, opts...)
		return err
	case client.DocKey:
		_, err := c.UpdateWithKey(ctx, t, updater, opts...)
		return err
	case []client.DocKey:
		_, err := c.UpdateWithKeys(ctx, t, updater, opts...)
		return err
	default:
		return ErrInvalidUpdateTarget
	}
}

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithFilter(ctx context.Context, filter interface{}, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithFilter(ctx, txn, filter, updater, opts...)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKey updates using a DocKey to target a single document for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKey(ctx context.Context, key client.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKey(ctx, txn, key, updater, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKeys is the same as UpdateWithKey but accepts multiple keys as a slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKeys(ctx context.Context, keys []client.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKeys(ctx, txn, keys, updater, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

func (c *collection) updateWithKey(ctx context.Context, txn datastore.Txn, key client.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isPatch = false
	default:
		return nil, ErrInvalidUpdater
	}

	doc, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	v, err := doc.ToMap()
	if err != nil {
		return nil, err
	}

	if isPatch {
		// todo
	} else {
		err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
	}
	if err != nil {
		return nil, err
	}

	results := &client.UpdateResult{
		Count:   1,
		DocKeys: []string{key.String()},
	}
	return results, nil
}

func (c *collection) updateWithKeys(ctx context.Context, txn datastore.Txn, keys []client.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isPatch = false
	default:
		return nil, ErrInvalidUpdater
	}

	results := &client.UpdateResult{
		DocKeys: make([]string, len(keys)),
	}
	for i, key := range keys {
		doc, err := c.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		v, err := doc.ToMap()
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else {
			err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, err
		}

		results.DocKeys[i] = key.String()
		results.Count++
	}
	return results, nil
}

func (c *collection) updateWithFilter(
	ctx context.Context,
	txn datastore.Txn,
	filter interface{},
	updater interface{},
	opts ...client.UpdateOpt) (*client.UpdateResult, error) {

	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	isMerge := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isMerge = true
	default:
		return nil, ErrInvalidUpdater
	}

	// scan through docs with filter
	query, err := c.makeSelectionQuery(ctx, txn, filter)
	if err != nil {
		return nil, err
	}
	if err = query.Start(); err != nil {
		return nil, err
	}

	// If the query object isn't properly closed at any exit point log the error.
	defer func() {
		if err := query.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close query after filter update", err)
		}
	}()

	results := &client.UpdateResult{
		DocKeys: make([]string, 0),
	}

	// loop while we still have results from the filter query
	for {
		next, nextErr := query.Next()
		if nextErr != nil {
			return nil, err
		}
		// if theres no more records from the query, jump out of the loop
		if !next {
			break
		}

		// Get the document, and apply the patch
		doc := query.Values()
		if isPatch {
			err = c.applyPatch(txn, doc, patch.([]map[string]interface{}))
		} else if isMerge { // else is fine here
			err = c.applyMerge(ctx, txn, doc, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, err
		}

		// add successful updated doc to results
		results.DocKeys = append(results.DocKeys, doc["_key"].(string))
		results.Count++
	}

	return results, nil
}

func (c *collection) applyPatch(txn datastore.Txn, doc map[string]interface{}, patch []map[string]interface{}) error {
	for _, op := range patch {
		path, ok := op["path"].(string)
		if !ok {
			return errors.New("Missing document field to update")
		}

		targetCollection, _, err := c.getCollectionForPatchOpPath(txn, path)
		if err != nil {
			return err
		}

		key, err := c.getTargetKeyForPatchPath(txn, doc, path)
		if err != nil {
			return err
		}
		field, val, _ := getValFromDocForPatchPath(doc, path)
		if err := targetCollection.applyPatchOp(txn, key, field, val, op); err != nil {
			return err
		}
	}

	// completed patch update
	return nil
}

func (c *collection) applyPatchOp(txn datastore.Txn, dockey string, field string, currentVal interface{}, patchOp map[string]interface{}) error {
	return nil
}

func (c *collection) applyMerge(ctx context.Context, txn datastore.Txn, doc map[string]interface{}, merge map[string]interface{}) error {
	keyStr, ok := doc["_key"].(string)
	if !ok {
		return errors.New("Document is missing key")
	}
	key := core.DataStoreKey{DocKey: keyStr}
	links := make([]core.DAGLink, 0)
	for mfield, mval := range merge {
		if _, ok := mval.(map[string]interface{}); ok {
			return ErrInvalidMergeValueType
		}

		fd, valid := c.desc.GetField(mfield)
		if !valid {
			return errors.New("Invalid field in Patch")
		}

		cval, err := validateFieldSchema(mval, fd)
		if err != nil {
			return err
		}

		// handle Int/Float case
		// JSON is annoying in that it represents all numbers
		// as Float64s. So our merge object contains float64s
		// even for fields defined as Ints, which causes issues
		// when we serialize that in CBOR. To generate the delta
		// payload.
		// So let's just make sure ints are ints ref: https://play.golang.org/p/djThEqGXtvR
		if fd.Kind == client.FieldKind_INT {
			merge[mfield] = int64(mval.(float64))
		}

		val := client.NewCBORValue(fd.Typ, cval)
		fieldKey := c.getFieldKey(key, mfield)
		c, err := c.saveDocValue(ctx, txn, c.getPrimaryIndexDocKey(fieldKey), val)
		if err != nil {
			return err
		}
		// links[mfield] = c
		links = append(links, core.DAGLink{
			Name: mfield,
			Cid:  c,
		})
	}

	// Update CompositeDAG
	em, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		return err
	}
	buf, err := em.Marshal(merge)
	if err != nil {
		return err
	}
	if _, err := c.saveValueToMerkleCRDT(ctx, txn, c.getPrimaryIndexDocKey(key), client.COMPOSITE, buf, links); err != nil {
		return err
	}

	// If this a a Batch masked as a Transaction
	// commit our writes so we can see them.
	// Batches don't maintain serializability, or
	// linearization, or any other transaction
	// semantics, which the user already knows
	// otherwise they wouldn't use a datastore
	// that doesn't support proper transactions.
	// So let's just commit, and keep going.
	// @todo: Change this on the Txn.BatchShim
	// structure
	if txn.IsBatch() {
		if err := txn.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldSchema takes a given value as an interface,
// and ensures it matches the supplied field description.
// It will do any minor parsing, like dates, and return
// the typed value again as an interface.
func validateFieldSchema(val interface{}, field client.FieldDescription) (interface{}, error) {
	var cval interface{}
	var err error
	var ok bool
	switch field.Kind {
	case client.FieldKind_DocKey, client.FieldKind_STRING:
		cval, ok = val.(string)
	case client.FieldKind_STRING_ARRAY:
		if val == nil {
			ok = true
			cval = nil
			break
		}
		untypedCollection := val.([]interface{})
		stringArray := make([]string, len(untypedCollection))
		for i, value := range untypedCollection {
			if value == nil {
				stringArray[i] = ""
				continue
			}
			stringArray[i], ok = value.(string)
			if !ok {
				return nil, fmt.Errorf("Failed to cast value: %v of type: %T to string", value, value)
			}
		}
		ok = true
		cval = stringArray
	case client.FieldKind_BOOL:
		cval, ok = val.(bool)
	case client.FieldKind_BOOL_ARRAY:
		if val == nil {
			ok = true
			cval = nil
			break
		}
		untypedCollection := val.([]interface{})
		boolArray := make([]bool, len(untypedCollection))
		for i, value := range untypedCollection {
			boolArray[i], ok = value.(bool)
			if !ok {
				return nil, fmt.Errorf("Failed to cast value: %v of type: %T to bool", value, value)
			}
		}
		ok = true
		cval = boolArray
	case client.FieldKind_FLOAT, client.FieldKind_DECIMAL:
		cval, ok = val.(float64)
	case client.FieldKind_FLOAT_ARRAY:
		if val == nil {
			ok = true
			cval = nil
			break
		}
		untypedCollection := val.([]interface{})
		floatArray := make([]float64, len(untypedCollection))
		for i, value := range untypedCollection {
			floatArray[i], ok = value.(float64)
			if !ok {
				return nil, fmt.Errorf("Failed to cast value: %v of type: %T to float64", value, value)
			}
		}
		ok = true
		cval = floatArray

	case client.FieldKind_DATE:
		var sval string
		sval, ok = val.(string)
		cval, err = time.Parse(time.RFC3339, sval)
	case client.FieldKind_INT:
		var fval float64
		fval, ok = val.(float64)
		if !ok {
			return nil, ErrInvalidMergeValueType
		}
		cval = int64(fval)
	case client.FieldKind_INT_ARRAY:
		if val == nil {
			ok = true
			cval = nil
			break
		}
		untypedCollection := val.([]interface{})
		intArray := make([]int64, len(untypedCollection))
		for i, value := range untypedCollection {
			valueAsFloat, castOk := value.(float64)
			if !castOk {
				return nil, fmt.Errorf("Failed to cast value: %v of type: %T to float64", value, value)
			}
			intArray[i] = int64(valueAsFloat)
		}
		ok = true
		cval = intArray
	case client.FieldKind_OBJECT, client.FieldKind_OBJECT_ARRAY,
		client.FieldKind_FOREIGN_OBJECT, client.FieldKind_FOREIGN_OBJECT_ARRAY:
		err = errors.New("Merge doesn't support sub types yet")
	}

	if !ok {
		return nil, ErrInvalidMergeValueType
	}
	if err != nil {
		return nil, err
	}

	return cval, err
}

func (c *collection) applyMergePatchOp( //nolint:unused
	txn datastore.Txn,
	docKey string,
	field string,
	currentVal interface{},
	targetVal interface{}) error {
	return nil
}

// makeQuery constructs a simple query of the collection using the given filter.
// currently it doesn't support any other query operation other than filters.
// (IE: No limit, order, etc)
// Additionally it only queries for the root scalar fields of the object
func (c *collection) makeSelectionQuery(
	ctx context.Context,
	txn datastore.Txn,
	filter interface{},
	opts ...client.UpdateOpt) (planner.Query, error) {
	var f *parser.Filter
	var err error
	switch fval := filter.(type) {
	case string:
		if fval == "" {
			return nil, errors.New("Invalid filter")
		}
		f, err = parser.NewFilterFromString(fval)
		if err != nil {
			return nil, err
		}
	case *parser.Filter:
		f = fval
	default:
		return nil, errors.New("Invalid filter")
	}
	if filter == "" {
		return nil, errors.New("Invalid filter")
	}
	slct, err := c.makeSelectLocal(f)
	if err != nil {
		return nil, err
	}

	return c.db.queryExecutor.MakeSelectQuery(ctx, c.db, txn, slct)
}

func (c *collection) makeSelectLocal(filter *parser.Filter) (*parser.Select, error) {
	slct := &parser.Select{
		Name:   c.Name(),
		Filter: filter,
		Fields: make([]parser.Selection, len(c.desc.Schema.Fields)),
	}

	for i, fd := range c.desc.Schema.Fields {
		if fd.IsObject() {
			continue
		}
		slct.Fields[i] = parser.Field{Name: fd.Name}
	}

	return slct, nil
}

// getTypeAndCollectionForPatch parses the Patch op path values
// and compares it against the collection schema.
// If it's within the schema, then patchIsSubType is false
// subTypeName is empty.
// If the target type is an array, isArray is true.
// May need to query the database for other schema types
// which requires a db transaction. It is recommended
// to use collection.WithTxn(txn) for this function call.
func (c *collection) getCollectionForPatchOpPath(txn datastore.Txn, path string) (col *collection, isArray bool, err error) {
	return nil, false, nil
}

// getTargetKeyForPatchPath walks through the given doc and Patch path.
// It returns the
func (c *collection) getTargetKeyForPatchPath(txn datastore.Txn, doc map[string]interface{}, path string) (string, error) {
	_, length := splitPatchPath(path)
	if length == 0 {
		return "", errors.New("Invalid patch op path")
	} else if length > 0 {

	}
	return "", nil
}

func splitPatchPath(path string) ([]string, int) {
	path = strings.TrimPrefix(path, "/")
	pathParts := strings.Split(path, "/")
	return pathParts, len(pathParts)
}

func getValFromDocForPatchPath(doc map[string]interface{}, path string) (string, interface{}, bool) {
	pathParts, length := splitPatchPath(path)
	if length == 0 {
		return "", nil, false
	}
	return getMapProp(doc, pathParts, length)
}

func getMapProp(doc map[string]interface{}, paths []string, length int) (string, interface{}, bool) {
	val, ok := doc[paths[0]]
	if !ok {
		return "", nil, false
	}
	if length > 1 {
		doc, ok := val.(map[string]interface{})
		if !ok {
			return "", nil, false
		}
		return getMapProp(doc, paths[1:], length-1)
	}
	return paths[0], val, true
}

type patcher interface{}

func parseUpdater(updater interface{}) (patcher, error) {
	switch v := updater.(type) {
	case string:
		return parseUpdaterString(v)
	case []interface{}:
		return parseUpdaterSlice(v)
	case []map[string]interface{}, map[string]interface{}:
		return patcher(v), nil
	case nil:
		return nil, ErrUpdateEmpty
	default:
		return nil, ErrInvalidUpdater
	}
}

func parseUpdaterString(v string) (patcher, error) {
	if v == "" {
		return nil, ErrUpdateEmpty
	}
	var i interface{}
	if err := json.Unmarshal([]byte(v), &i); err != nil {
		return nil, err
	}
	return parseUpdater(i)
}

// converts an []interface{} to []map[string]interface{}
// which is required to be an array of Patch Ops
func parseUpdaterSlice(v []interface{}) (patcher, error) {
	if len(v) == 0 {
		return nil, ErrUpdateEmpty
	}

	patches := make([]map[string]interface{}, len(v))
	for i, patch := range v {
		p, ok := patch.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidUpdater
		}
		patches[i] = p
	}

	return parseUpdater(patches)
}

/*

filter := NewFilterFromString("Name: {_eq: 'bob'}")

filter := db.NewQuery().And()

*/
