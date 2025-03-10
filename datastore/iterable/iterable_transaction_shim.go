// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package iterable

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
)

// implement interface check
var _ Iterable = (*iterableShim)(nil)
var _ Iterator = (*iteratorShim)(nil)

type iterableShim struct {
	ds.Read
}

type iteratorShim struct {
	readable iterableShim
	results  dsq.Results
	q        dsq.Query
}

func NewIterable(readable ds.Read) Iterable {
	return iterableShim{
		readable,
	}
}

func (shim iterableShim) GetIterator(q dsq.Query) (Iterator, error) {
	return &iteratorShim{
		readable: shim,
		q:        q,
	}, nil
}

func (shim *iteratorShim) IteratePrefix(
	ctx context.Context,
	startPrefix ds.Key,
	endPrefix ds.Key,
) (dsq.Results, error) {
	if shim.results != nil {
		err := shim.results.Close()
		if err != nil {
			return nil, err
		}
	}

	query := shim.q
	// If the prefix range only covers one prefix then we don't have to do the
	// horrible work-around in the else clause
	if prefixEnd(startPrefix) == endPrefix {
		query.Prefix = startPrefix.String()
		results, err := shim.readable.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		shim.results = results
	} else {
		startBytes := startPrefix.Bytes()
		endBytes := endPrefix.Bytes()
		lastSharedIndex := 0
		for i := 0; i < len(startBytes) && i < len(endBytes); i++ {
			if startBytes[i] != endBytes[i] {
				break
			}
			lastSharedIndex += 1
		}
		query.Prefix = string(startBytes[:lastSharedIndex])
		query.Filters = append(query.Filters, betweenFilter{
			start: startPrefix.String(),
			end:   endPrefix.String(),
		})
		results, err := shim.readable.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		shim.results = results
	}
	return shim.results, nil
}

type betweenFilter struct {
	start string
	end   string
}

func (f betweenFilter) Filter(e dsq.Entry) bool {
	return e.Key >= f.start && e.Key <= f.end
}

func (shim *iteratorShim) Close() error {
	if shim.results == nil {
		return nil
	}
	return shim.results.Close()
}

var keyMax = string(([]byte{0xff, 0xff}))

// PrefixEnd determines the end key given key as a prefix, that is the
// key that sorts precisely behind all keys starting with prefix: "1"
// is added to the final byte and the carry propagated. The special
// cases of nil and KeyMin always returns KeyMax.
func prefixEnd(k ds.Key) ds.Key {
	if len(k.Bytes()) == 0 {
		return ds.NewKey(keyMax)
	}
	return ds.NewKey(string(bytesPrefixEnd(k.Bytes())))
}

func bytesPrefixEnd(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	// This statement will only be reached if the key is already a
	// maximal byte string (i.e. already \xff...).
	return b
}
