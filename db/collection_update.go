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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/planner"

	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"

	cbor "github.com/fxamacker/cbor/v2"
	"github.com/valyala/fastjson"
)

var (
	ErrUpdateTargetEmpty     = errors.New("The doc update targeter cannot be empty")
	ErrUpdateEmpty           = errors.New("The doc update cannot be empty")
	ErrInvalidMergeValueType = errors.New(
		"The type of value in the merge patch doesn't match the schema",
	)
)

// UpdateWith updates a target document using the given updater type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Update.
// Eg: UpdateWithFilter or UpdateWithKey
func (c *collection) UpdateWith(
	ctx context.Context,
	target interface{},
	updater string,
) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]interface{}, *parser.Filter:
		return c.UpdateWithFilter(ctx, t, updater)
	case client.DocKey:
		return c.UpdateWithKey(ctx, t, updater)
	case []client.DocKey:
		return c.UpdateWithKeys(ctx, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithFilter(
	ctx context.Context,
	filter interface{},
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithFilter(ctx, txn, filter, updater)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKey updates using a DocKey to target a single document for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKey(
	ctx context.Context,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKey(ctx, txn, key, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKeys is the same as UpdateWithKey but accepts multiple keys as a slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKeys(
	ctx context.Context,
	keys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKeys(ctx, txn, keys, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

func (c *collection) updateWithKey(
	ctx context.Context,
	txn datastore.Txn,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
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
		err = c.applyMerge(ctx, txn, v, parsedUpdater.GetObject())
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

func (c *collection) updateWithKeys(
	ctx context.Context,
	txn datastore.Txn,
	keys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
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
			err = c.applyMerge(ctx, txn, v, parsedUpdater.GetObject())
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
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	isMerge := false
	switch parsedUpdater.Type() {
	case fastjson.TypeArray:
		isPatch = true
	case fastjson.TypeObject:
		isMerge = true
	default:
		return nil, client.ErrInvalidUpdater
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

	docMap := query.DocumentMap()

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
		doc := docMap.ToMap(query.Value())
		if isPatch {
			err = c.applyPatch(txn, doc, parsedUpdater.GetArray())
		} else if isMerge { // else is fine here
			err = c.applyMerge(ctx, txn, doc, parsedUpdater.GetObject())
		}
		if err != nil {
			return nil, err
		}

		// add successful updated doc to results
		results.DocKeys = append(results.DocKeys, doc[parserTypes.DocKeyFieldName].(string))
		results.Count++
	}

	return results, nil
}

func (c *collection) applyPatch(
	txn datastore.Txn,
	doc map[string]interface{},
	patch []*fastjson.Value,
) error {
	for _, op := range patch {
		opObject, err := op.Object()
		if err != nil {
			return err
		}
		path, err := opObject.Get("path").StringBytes()
		if err != nil {
			return fmt.Errorf("missing document field to update: %w", err)
		}

		targetCollection, _, err := c.getCollectionForPatchOpPath(txn, string(path))
		if err != nil {
			return err
		}

		key, err := c.getTargetKeyForPatchPath(txn, doc, string(path))
		if err != nil {
			return err
		}
		field, val, _ := getValFromDocForPatchPath(doc, string(path))
		if err := targetCollection.applyPatchOp(txn, key, field, val, opObject); err != nil {
			return err
		}
	}

	// completed patch update
	return nil
}

func (c *collection) applyPatchOp(
	txn datastore.Txn,
	dockey string,
	field string,
	currentVal interface{},
	patchOp *fastjson.Object,
) error {
	return nil
}

func (c *collection) applyMerge(
	ctx context.Context,
	txn datastore.Txn,
	doc map[string]interface{},
	merge *fastjson.Object,
) error {
	keyStr, ok := doc["_key"].(string)
	if !ok {
		return errors.New("document is missing key")
	}
	key := c.getPrimaryKey(keyStr)
	links := make([]core.DAGLink, 0)

	mergeMap := make(map[string]*fastjson.Value)
	merge.Visit(func(k []byte, v *fastjson.Value) {
		mergeMap[string(k)] = v
	})

	mergeCBOR := make(map[string]any)

	for mfield, mval := range mergeMap {
		if mval.Type() == fastjson.TypeObject {
			return ErrInvalidMergeValueType
		}

		fd, valid := c.desc.GetField(mfield)
		if !valid {
			return errors.New("invalid field in Patch")
		}

		cborVal, err := validateFieldSchema(mval, fd)
		if err != nil {
			return err
		}
		mergeCBOR[mfield] = cborVal

		val := client.NewCBORValue(fd.Typ, cborVal)
		fieldKey, fieldExists := c.tryGetFieldKey(key, mfield)
		if !fieldExists {
			return client.ErrFieldNotExist
		}

		c, err := c.saveDocValue(ctx, txn, fieldKey, val)
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
	buf, err := em.Marshal(mergeCBOR)
	if err != nil {
		return err
	}
	if _, err := c.saveValueToMerkleCRDT(
		ctx,
		txn,
		key.ToDataStoreKey(),
		client.COMPOSITE,
		buf,
		links,
	); err != nil {
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
func validateFieldSchema(val *fastjson.Value, field client.FieldDescription) (interface{}, error) {
	switch field.Kind {
	case client.FieldKind_DocKey, client.FieldKind_STRING:
		return getString(val)

	case client.FieldKind_STRING_ARRAY:
		return getArray(val, getString)

	case client.FieldKind_NILLABLE_STRING_ARRAY:
		return getNillableArray(val, getString)

	case client.FieldKind_BOOL:
		return getBool(val)

	case client.FieldKind_BOOL_ARRAY:
		return getArray(val, getBool)

	case client.FieldKind_NILLABLE_BOOL_ARRAY:
		return getNillableArray(val, getBool)

	case client.FieldKind_FLOAT, client.FieldKind_DECIMAL:
		return getFloat64(val)

	case client.FieldKind_FLOAT_ARRAY:
		return getArray(val, getFloat64)

	case client.FieldKind_NILLABLE_FLOAT_ARRAY:
		return getNillableArray(val, getFloat64)

	case client.FieldKind_DATE:
		return getDate(val)

	case client.FieldKind_INT:
		return getInt64(val)

	case client.FieldKind_INT_ARRAY:
		return getArray(val, getInt64)

	case client.FieldKind_NILLABLE_INT_ARRAY:
		return getNillableArray(val, getInt64)

	case client.FieldKind_OBJECT, client.FieldKind_OBJECT_ARRAY,
		client.FieldKind_FOREIGN_OBJECT, client.FieldKind_FOREIGN_OBJECT_ARRAY:
		return nil, errors.New("merge doesn't support sub types yet")
	}

	return nil, errors.New("unsupported field kind")
}

func getString(v *fastjson.Value) (string, error) {
	b, err := v.StringBytes()
	return string(b), err
}

func getBool(v *fastjson.Value) (bool, error) {
	return v.Bool()
}

func getFloat64(v *fastjson.Value) (float64, error) {
	return v.Float64()
}

func getInt64(v *fastjson.Value) (int64, error) {
	return v.Int64()
}

func getDate(v *fastjson.Value) (time.Time, error) {
	s, err := getString(v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, s)
}

func getArray[T any](
	val *fastjson.Value,
	typeGetter func(*fastjson.Value) (T, error),
) ([]T, error) {
	if val.Type() == fastjson.TypeNull {
		return nil, nil
	}

	valArray, err := val.Array()
	if err != nil {
		return nil, err
	}

	arr := make([]T, len(valArray))
	for i, arrItem := range valArray {
		if arrItem.Type() == fastjson.TypeNull {
			continue
		}
		arr[i], err = typeGetter(arrItem)
		if err != nil {
			return nil, err
		}
	}

	return arr, nil
}

func getNillableArray[T any](
	val *fastjson.Value,
	typeGetter func(*fastjson.Value) (T, error),
) ([]*T, error) {
	if val.Type() == fastjson.TypeNull {
		return nil, nil
	}

	valArray, err := val.Array()
	if err != nil {
		return nil, err
	}

	arr := make([]*T, len(valArray))
	for i, arrItem := range valArray {
		if arrItem.Type() == fastjson.TypeNull {
			continue
		}
		v, err := typeGetter(arrItem)
		if err != nil {
			return nil, err
		}
		arr[i] = &v
	}

	return arr, nil
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
) (planner.Query, error) {
	mapping := c.createMapping()
	var f *mapper.Filter
	var err error
	switch fval := filter.(type) {
	case string:
		if fval == "" {
			return nil, errors.New("Invalid filter")
		}
		var p *parser.Filter
		p, err = parser.NewFilterFromString(fval)
		if err != nil {
			return nil, err
		}
		f = mapper.ToFilter(p, mapping)
	case *mapper.Filter:
		f = fval
	default:
		return nil, errors.New("Invalid filter")
	}
	if filter == "" {
		return nil, errors.New("Invalid filter")
	}
	slct, err := c.makeSelectLocal(f, mapping)
	if err != nil {
		return nil, err
	}

	return c.db.queryExecutor.MakeSelectQuery(ctx, c.db, txn, slct)
}

func (c *collection) makeSelectLocal(filter *mapper.Filter, mapping *core.DocumentMapping) (*mapper.Select, error) {
	slct := &mapper.Select{
		Targetable: mapper.Targetable{
			Field: mapper.Field{
				Name: c.Name(),
			},
			Filter: filter,
		},
		Fields:          make([]mapper.Requestable, len(c.desc.Schema.Fields)),
		DocumentMapping: *mapping,
	}

	for _, fd := range c.Schema().Fields {
		if fd.IsObject() {
			continue
		}
		index := int(fd.ID)
		slct.Fields = append(slct.Fields, &mapper.Field{
			Index: index,
			Name:  fd.Name,
		})
	}

	return slct, nil
}

func (c *collection) createMapping() *core.DocumentMapping {
	mapping := core.NewDocumentMapping()
	mapping.Add(core.DocKeyFieldIndex, parserTypes.DocKeyFieldName)
	for _, fd := range c.Schema().Fields {
		if fd.IsObject() {
			continue
		}
		index := int(fd.ID)
		mapping.Add(index, fd.Name)
		mapping.RenderKeys = append(mapping.RenderKeys,
			core.RenderKey{
				Index: index,
				Key:   fd.Name,
			},
		)
	}
	return mapping
}

// getTypeAndCollectionForPatch parses the Patch op path values
// and compares it against the collection schema.
// If it's within the schema, then patchIsSubType is false
// subTypeName is empty.
// If the target type is an array, isArray is true.
// May need to query the database for other schema types
// which requires a db transaction. It is recommended
// to use collection.WithTxn(txn) for this function call.
func (c *collection) getCollectionForPatchOpPath(
	txn datastore.Txn,
	path string,
) (col *collection, isArray bool, err error) {
	return nil, false, nil
}

// getTargetKeyForPatchPath walks through the given doc and Patch path.
// It returns the
func (c *collection) getTargetKeyForPatchPath(
	txn datastore.Txn,
	doc map[string]interface{},
	path string,
) (string, error) {
	_, length := splitPatchPath(path)
	if length == 0 {
		return "", errors.New("Invalid patch op path")
	}

	return "", nil
}

func splitPatchPath(path string) ([]string, int) {
	path = strings.TrimPrefix(path, "/")
	pathParts := strings.Split(path, "/")
	return pathParts, len(pathParts)
}

func getValFromDocForPatchPath(
	doc map[string]interface{},
	path string,
) (string, interface{}, bool) {
	pathParts, length := splitPatchPath(path)
	if length == 0 {
		return "", nil, false
	}
	return getMapProp(doc, pathParts, length)
}

func getMapProp(
	doc map[string]interface{},
	paths []string,
	length int,
) (string, interface{}, bool) {
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

/*

filter := NewFilterFromString("Name: {_eq: 'bob'}")

filter := db.NewQuery().And()

*/
