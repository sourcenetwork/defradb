// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

// FullTextIndexPart identifies one logical record family in a full-text index.
type FullTextIndexPart byte

const (
	FullTextPostingPart FullTextIndexPart = 't'
	FullTextLengthPart  FullTextIndexPart = 'd'
	FullTextTotalsPart  FullTextIndexPart = 's'
)

// FullTextIndexKey addresses BM25 postings, document lengths, and corpus totals beneath the common
// (collection, index, epoch) secondary-index prefix.
//
// Layout:
//
//	/<collection>/<index>/<epoch>/t/<encoded-term>/<docShortID>
//	/<collection>/<index>/<epoch>/d/<docShortID>
//	/<collection>/<index>/<epoch>/s
type FullTextIndexKey struct {
	CollectionShortID uint32
	IndexID           uint32
	Epoch             uint32
	Part              FullTextIndexPart
	Term              string
	DocShortID        uint64
	offset            uint64
}

var _ Walkable = (*FullTextIndexKey)(nil)
var _ CollectionedKey = (*FullTextIndexKey)(nil)

func NewFullTextPostingKey(
	collectionShortID, indexID, epoch uint32,
	term string,
	docShortID uint64,
) FullTextIndexKey {
	return FullTextIndexKey{
		CollectionShortID: collectionShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Part:              FullTextPostingPart,
		Term:              term,
		DocShortID:        docShortID,
	}
}

func NewFullTextLengthKey(
	collectionShortID, indexID, epoch uint32,
	docShortID uint64,
) FullTextIndexKey {
	return FullTextIndexKey{
		CollectionShortID: collectionShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Part:              FullTextLengthPart,
		DocShortID:        docShortID,
	}
}

func NewFullTextTotalsKey(collectionShortID, indexID, epoch uint32) FullTextIndexKey {
	return FullTextIndexKey{
		CollectionShortID: collectionShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Part:              FullTextTotalsPart,
	}
}

func (k *FullTextIndexKey) Bytes() []byte {
	b := encoding.EncodeUvarintAscending([]byte{'/'}, uint64(k.CollectionShortID))
	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.IndexID))
	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.Epoch))
	b = append(b, '/', byte(k.Part))

	if k.Part == FullTextPostingPart && k.Term != "" {
		b = append(b, '/')
		b = encoding.EncodeStringAscending(b, k.Term)
	}
	if k.DocShortID != 0 {
		b = append(b, '/')
		b = append(b, EncodeDocShortID(k.DocShortID)...)
	}
	for i := uint64(0); i < k.offset; i++ {
		b = bytesPrefixEnd(b)
	}
	return b
}

func (k *FullTextIndexKey) ToString() string { return string(k.Bytes()) }
func (k *FullTextIndexKey) ToDS() ds.Key     { return ds.NewKey(k.ToString()) }

func (k *FullTextIndexKey) GetCollectionShortID() uint32 { return k.CollectionShortID }

func (k *FullTextIndexKey) PrefixEnd() Walkable {
	copy := *k
	copy.offset++
	return &copy
}
