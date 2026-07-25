// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encoding"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// The parts of a BM25 index's keyspace. Every key it writes is an ordinary IndexDataStoreKey
// under the index's own prefix, /<collection short ID>/<index ID>/<epoch>/, with one of these
// as the first field value. Staying inside that prefix is what leaves epoch handling,
// rebuild-on-version-switch and the garbage collector's range delete working unchanged:
//
//	<prefix>/t/<term>/<doc short ID> -> how often the term occurs in the document
//	<prefix>/d/<doc short ID>        -> the document's field length, in terms
//	<prefix>/s                       -> the collection's document count and summed field length
//
// There is no per-term document frequency record. It is the number of entries under
// <prefix>/t/<term>/, which a query reads as a range while it walks that term's documents
// anyway, so storing it would mean a read-modify-write per distinct term of every document
// written, on keys that the most common terms make hot across every writer.
const (
	bm25PostingPart = "t"
	bm25LengthPart  = "d"
	bm25TotalsPart  = "s"
)

// The scoring parameters of a BM25 index, settable through the index's options. k1 controls how
// quickly a repeated term stops adding to a document's score, and b how strongly a field longer
// than average is penalised. Only scoring reads them, so nothing here uses them beyond checking
// that what was given is usable.
const (
	bm25OptionK1  = "k1"
	bm25OptionB   = "b"
	bm25DefaultK1 = 1.2
	bm25DefaultB  = 0.75
)

// collectionBM25Index maintains everything needed to score a document against a query: how often
// each term occurs in it, how long its indexed field is, and the collection totals those are
// scored relative to.
//
// It does not reuse FieldIndexGenerator, which the trigram kind is built from. A generator emits
// the values to turn into keys and nothing else, so it can neither give an entry a value nor
// maintain anything per document or per collection, and all three are the point here.
type collectionBM25Index struct {
	collectionBaseIndex
}

var _ client.CollectionIndex = (*collectionBM25Index)(nil)

func init() {
	registerIndexKind(client.IndexKindBM25, func(base collectionBaseIndex) client.CollectionIndex {
		return &collectionBM25Index{collectionBaseIndex: base}
	})
}

// Save writes one entry per distinct term of the document's indexed field holding the term's
// frequency, records the field length, and folds the document into the collection totals.
//
// Saving a document the index already holds rewrites its entries and moves the totals by the
// difference in field length rather than counting the document again, which is what lets a
// backfill and a concurrent live write of the same document both run.
func (index *collectionBM25Index) Save(ctx context.Context, doc *client.Document) error {
	tokens, err := index.docTokens(doc)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		// No terms means no entries and nothing to count, so a document with no words in the
		// indexed field is simply not in the index. Nothing is lost: it can not match a query.
		return nil
	}

	collectionShortID, docShortID, err := index.docShortIDs(ctx, doc)
	if err != nil {
		return err
	}
	ds := datastore.CtxMustGetTxn(ctx).Datastore()

	var length uint64
	for term, frequency := range tokens {
		postingKey := index.key(collectionShortID, bm25PostingPart, term, docShortID)
		if err := ds.Set(ctx, &postingKey, encodeUvarints(uint64(frequency))); err != nil {
			return NewErrStoreIndexKey(err)
		}
		length += uint64(frequency)
	}

	lengthKey := index.key(collectionShortID, bm25LengthPart, "", docShortID)
	stored, alreadyIndexed, err := index.readUvarints(ctx, &lengthKey, 1)
	if err != nil {
		return err
	}
	if err := ds.Set(ctx, &lengthKey, encodeUvarints(length)); err != nil {
		return NewErrStoreIndexKey(err)
	}

	if alreadyIndexed {
		return index.addToTotals(ctx, collectionShortID, 0, int64(length)-int64(stored[0]))
	}
	return index.addToTotals(ctx, collectionShortID, 1, int64(length))
}

// Delete removes the document's entries and takes it back out of the collection totals. The
// terms it removes are recomputed from the document being deleted, so the caller must pass the
// document as it was indexed.
func (index *collectionBM25Index) Delete(ctx context.Context, doc *client.Document) error {
	tokens, err := index.docTokens(doc)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		// The document has no terms, so Save gave it no entries and did not count it.
		return nil
	}

	collectionShortID, docShortID, err := index.docShortIDs(ctx, doc)
	if err != nil {
		return err
	}

	lengthKey := index.key(collectionShortID, bm25LengthPart, "", docShortID)
	stored, indexed, err := index.readUvarints(ctx, &lengthKey, 1)
	if err != nil {
		return err
	}
	if !indexed {
		// During backfill, documents not yet reached have no entries, so we skip silently.
		if index.building {
			return nil
		}
		return NewErrCorruptedIndex(index.desc.Name)
	}

	for term := range tokens {
		if err := index.deleteIndexKey(ctx, index.key(collectionShortID, bm25PostingPart, term, docShortID)); err != nil {
			return err
		}
	}
	if err := datastore.CtxMustGetTxn(ctx).Datastore().Delete(ctx, &lengthKey); err != nil {
		return NewErrDeleteIndexKey(err)
	}
	return index.addToTotals(ctx, collectionShortID, -1, -int64(stored[0]))
}

func (index *collectionBM25Index) Update(ctx context.Context, oldDoc *client.Document, newDoc *client.Document) error {
	// An update that left the indexed field alone would rewrite every entry to the value it
	// already holds, so there is nothing to do.
	if !isUpdatingIndexedFields(index, oldDoc, newDoc) {
		return nil
	}
	if err := index.Delete(ctx, oldDoc); err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}
	if err := index.Save(ctx, newDoc); err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}
	return nil
}

// addToTotals applies a change to the index's document count and summed field length, from which
// the average field length that scoring normalises against is derived.
//
// ponytail: one key per index, read and written by every write to the collection, so concurrent
// writers conflict on it and retry. It is maintained exactly because the alternative for these
// two numbers is counting every document in the collection on every query, which is worse. If
// the write ceiling is reached before that read cost matters, split this record over a fixed
// number of keys chosen by document short ID and sum them on the read side: same read, and the
// contention divided by the number of shards.
//
// A total that has drifted below zero is clamped rather than wrapped. These are ranking inputs
// and a stale one only shifts scores, but a wrapped document count would make the inverse
// document frequency of every term meaningless.
func (index *collectionBM25Index) addToTotals(
	ctx context.Context,
	collectionShortID uint32,
	documents int64,
	length int64,
) error {
	key := index.key(collectionShortID, bm25TotalsPart, "", 0)
	totals, found, err := index.readUvarints(ctx, &key, 2)
	if err != nil {
		return err
	}
	if !found {
		totals = []uint64{0, 0}
	}
	value := encodeUvarints(addClamped(totals[0], documents), addClamped(totals[1], length))
	if err := datastore.CtxMustGetTxn(ctx).Datastore().Set(ctx, &key, value); err != nil {
		return NewErrStoreIndexKey(err)
	}
	return nil
}

// key returns the key of one part of the index's keyspace. term is empty for the parts that are
// not per-term, and docShortID is zero for the parts that are not per-document.
func (index *collectionBM25Index) key(
	collectionShortID uint32,
	part string,
	term string,
	docShortID uint64,
) keys.IndexDataStoreKey {
	fields := []keys.IndexedField{{Value: client.NewNormalString(part)}}
	if term != "" {
		fields = append(fields, keys.IndexedField{Value: client.NewNormalString(term)})
	}
	key := keys.NewIndexDataStoreKey(collectionShortID, index.desc.ID, index.epoch, fields)
	key.DocShortID = docShortID
	return key
}

// docShortIDs resolves the collection and document short IDs the index's keys are built from.
func (index *collectionBM25Index) docShortIDs(
	ctx context.Context,
	doc *client.Document,
) (uint32, uint64, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, index.collection.Version().CollectionID)
	if err != nil {
		return 0, 0, err
	}
	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
	if err != nil {
		return 0, 0, err
	}
	if !found {
		return 0, 0, client.ErrDocumentNotFoundOrNotAuthorized
	}
	return collectionShortID, docShortID, nil
}

// docTokens returns the terms of the document's indexed field and how often each occurs. A BM25
// index is on exactly one string field, which is checked when it is created.
func (index *collectionBM25Index) docTokens(doc *client.Document) (map[string]int, error) {
	values, err := index.getDocFieldValues(doc)
	if err != nil {
		return nil, err
	}
	str, ok := values[0].String()
	if !ok {
		nillableStr, isString := values[0].NillableString()
		if !isString || !nillableStr.HasValue() {
			return nil, nil
		}
		str = nillableStr.Value()
	}
	return core.Tokens(str), nil
}

// readUvarints reads count uvarints from the value stored at key, reporting false when the key
// does not exist.
func (index *collectionBM25Index) readUvarints(
	ctx context.Context,
	key *keys.IndexDataStoreKey,
	count int,
) ([]uint64, bool, error) {
	data, err := datastore.CtxMustGetTxn(ctx).Datastore().Get(ctx, key)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, NewErrCheckIndexKeyExists(err, index.desc.Name)
	}
	values := make([]uint64, count)
	for i := range values {
		data, values[i], err = encoding.DecodeUvarintAscending(data)
		if err != nil {
			return nil, false, NewErrCorruptedIndex(index.desc.Name)
		}
	}
	return values, true, nil
}

func encodeUvarints(values ...uint64) []byte {
	var encoded []byte
	for _, value := range values {
		encoded = encoding.EncodeUvarintAscending(encoded, value)
	}
	return encoded
}

// addClamped adds a signed change to an unsigned total, stopping at zero rather than wrapping.
func addClamped(total uint64, change int64) uint64 {
	if change >= 0 {
		return total + uint64(change)
	}
	if uint64(-change) > total {
		return 0
	}
	return total - uint64(-change)
}

// validateBM25Index checks a new BM25 index request against the collection.
func validateBM25Index(def client.CollectionVersion, req client.NewIndexRequest) error {
	if req.Unique {
		return ErrBM25IndexUnique
	}
	if len(req.Fields) != 1 {
		return ErrBM25IndexNotSingleField
	}
	field, found := def.GetFieldByName(req.Fields[0].Name)
	if !found {
		return NewErrNonExistingFieldForIndex(req.Fields[0].Name)
	}
	if field.Kind != client.FieldKind_STRING && field.Kind != client.FieldKind_NILLABLE_STRING {
		return NewErrBM25IndexFieldNotString(field.Name, field.Kind.String())
	}
	for name := range req.Options {
		if name != bm25OptionK1 && name != bm25OptionB {
			return NewErrBM25IndexUnknownOption(name)
		}
	}
	k1, err := bm25Option(req.Options, bm25OptionK1, bm25DefaultK1)
	if err != nil {
		return err
	}
	if k1 < 0 {
		return NewErrBM25IndexInvalidOption(bm25OptionK1, k1)
	}
	b, err := bm25Option(req.Options, bm25OptionB, bm25DefaultB)
	if err != nil {
		return err
	}
	if b < 0 || b > 1 {
		return NewErrBM25IndexInvalidOption(bm25OptionB, b)
	}
	return nil
}

// bm25Option reads a numeric index option, returning fallback when it is not set.
//
// An integer written in a GraphQL literal parses to an int, while the same description read back
// from the JSON it is stored as gives a float64, so both are accepted. For the same reason a
// description parsed from a schema and one read back from the store can differ in the Go types
// they hold, and so can not be compared with reflect.DeepEqual.
func bm25Option(options map[string]any, name string, fallback float64) (float64, error) {
	value, ok := options[name]
	if !ok {
		return fallback, nil
	}
	switch number := value.(type) {
	case float64:
		return number, nil
	case int:
		return float64(number), nil
	default:
		return 0, NewErrBM25IndexInvalidOption(name, value)
	}
}
