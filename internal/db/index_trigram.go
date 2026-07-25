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
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

// trigrams returns the distinct trigrams of val: the value lowercased, then every 3-byte
// window of it, stride 1 and overlapping, with no padding.
//
// The windows are bytes, not runes. A multi-byte rune is therefore split across windows,
// which is safe because a trigram index only ever produces candidates and the real pattern
// match is re-run on each one.
//
// A value shorter than 3 bytes produces no trigrams at all, so the document holding it has
// no entry in the index and cannot be found through it. That is intended: padding short
// values to reach three bytes would put every one of them under the same handful of keys.
func trigrams(val string) []string {
	val = strings.ToLower(val)
	if len(val) < 3 {
		return nil
	}
	// One entry per distinct trigram per document. A repeated trigram would otherwise write
	// the same key twice, and the second delete of it would report the index as corrupt.
	seen := make(map[string]struct{}, len(val)-2)
	result := make([]string, 0, len(val)-2)
	for i := 0; i+3 <= len(val); i++ {
		trigram := val[i : i+3]
		if _, ok := seen[trigram]; ok {
			continue
		}
		seen[trigram] = struct{}{}
		result = append(result, trigram)
	}
	return result
}

// TrigramFieldGenerator fans a string field value out into one index entry per distinct
// trigram of it. This is the same shape as ArrayFieldGenerator - one field value, many
// entries - so the trigram index is the ordinary non-unique index over these values.
type TrigramFieldGenerator struct{}

func (g *TrigramFieldGenerator) Generate(value client.NormalValue, f func(client.NormalValue) error) error {
	str, ok := value.String()
	if !ok {
		nillableStr, isString := value.NillableString()
		if !isString || !nillableStr.HasValue() {
			// A nil value has no trigrams, so it gets no entry. Nothing is lost: a trigram index
			// only serves substring and pattern matching, and neither matches a nil value.
			return nil
		}
		str = nillableStr.Value()
	}

	for _, trigram := range trigrams(str) {
		if err := f(client.NewNormalString(trigram)); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	// A trigram index writes one ordinary key per (trigram, document) with an empty value, which
	// is exactly what the non-unique ordered index does once the generator has fanned the value
	// out. It therefore needs no implementation of its own.
	registerIndexKind(client.IndexKindTrigram, func(base collectionBaseIndex) client.CollectionIndex {
		return &collectionSimpleIndex{collectionBaseIndex: base}
	})
}

// validateTrigramIndex checks a new trigram index request against the collection.
func validateTrigramIndex(def client.CollectionVersion, req client.NewIndexRequest) error {
	if req.Unique {
		return ErrTrigramIndexUnique
	}
	if len(req.Fields) != 1 {
		return ErrTrigramIndexNotSingleField
	}
	field, found := def.GetFieldByName(req.Fields[0].Name)
	if !found {
		return NewErrNonExistingFieldForIndex(req.Fields[0].Name)
	}
	if field.Kind != client.FieldKind_STRING && field.Kind != client.FieldKind_NILLABLE_STRING {
		return NewErrTrigramIndexFieldNotString(field.Name, field.Kind.String())
	}
	return nil
}
