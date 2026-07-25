// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"github.com/sourcenetwork/defradb/client"
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
//
// The layout lives here rather than beside the index for the same reason [Tokens] does: the
// write path (internal/db) and the read path (internal/db/fetcher) must build the same keys,
// and internal/db imports internal/db/fetcher.
const (
	BM25PostingPart = "t"
	BM25LengthPart  = "d"
	BM25TotalsPart  = "s"
)

// The scoring parameters of a BM25 index, settable through the index's options. k1 controls how
// quickly a repeated term stops adding to a document's score, and b how strongly a field longer
// than average is penalised.
const (
	BM25OptionK1  = "k1"
	BM25OptionB   = "b"
	BM25DefaultK1 = 1.2
	BM25DefaultB  = 0.75
)

// BM25Key returns the key of one part of a BM25 index's keyspace. term is empty for the parts
// that are not per-term, and docShortID is zero for the parts that are not per-document.
func BM25Key(
	collectionShortID, indexID, epoch uint32,
	part string,
	term string,
	docShortID uint64,
) keys.IndexDataStoreKey {
	fields := []keys.IndexedField{{Value: client.NewNormalString(part)}}
	if term != "" {
		fields = append(fields, keys.IndexedField{Value: client.NewNormalString(term)})
	}
	key := keys.NewIndexDataStoreKey(collectionShortID, indexID, epoch, fields)
	key.DocShortID = docShortID
	return key
}

// BM25Option reads a numeric index option, returning fallback when it is not set and false when
// what is set is not a number.
//
// An integer written in a GraphQL literal parses to an int, while the same description read back
// from the JSON it is stored as gives a float64, so both are accepted. For the same reason a
// description parsed from a schema and one read back from the store can differ in the Go types
// they hold, and so can not be compared with reflect.DeepEqual.
func BM25Option(options map[string]any, name string, fallback float64) (float64, bool) {
	value, ok := options[name]
	if !ok {
		return fallback, true
	}
	switch number := value.(type) {
	case float64:
		return number, true
	case int:
		return float64(number), true
	default:
		return 0, false
	}
}
