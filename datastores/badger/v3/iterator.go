// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package badger

// this is quite similar in principle to https://github.com/MikkelHJuul/bIter/blob/main/iterator.go that John linked - maybe just use/wrap that
import (
	"context"
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	goprocess "github.com/jbenet/goprocess"
	"github.com/sourcenetwork/defradb/datastores/iterable"
)

type BadgerIterator struct {
	iterator       *badger.Iterator
	resultsBuilder *dsq.ResultBuilder
	query          dsq.Query
	txn            txn
	skipped        int
	sent           int
	closedEarly    bool
	iteratorLock   sync.RWMutex
}

func (t *txn) GetIterator(q dsq.Query) (iterable.Iterator, error) {
	opt := badger.DefaultIteratorOptions
	// Prefetching prevents the re-use of the iterator
	opt.PrefetchValues = false

	// Handle ordering
	if len(q.Orders) > 0 {
		switch orderType := q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
		// We order by key by default.
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// Reverse order by key
			opt.Reverse = true
		default:
			return nil, fmt.Errorf("Order format not supported: %v", orderType)
		}
	}

	badgerIterator := t.txn.NewIterator(opt)

	iterator := BadgerIterator{
		iterator: badgerIterator,
		txn:      *t,
	}

	return &iterator, nil
}

func (iterator *BadgerIterator) Close() error {
	// There is a race condition between `iterator.iterator.Next()` and `iterator.iterator.Close()` which we have to protect against here
	iterator.iteratorLock.Lock()
	iterator.iterator.Close()
	iterator.iteratorLock.Unlock()
	return nil
}

func (iterator *BadgerIterator) next() {
	// There is a race condition between `iterator.iterator.Next()` and `iterator.iterator.Close()` which we have to protect against here
	iterator.iteratorLock.RLock()
	iterator.iterator.Next()
	iterator.iteratorLock.RUnlock()
}

func (iterator *BadgerIterator) IteratePrefix(ctx context.Context, prefix ds.Key) (dsq.Results, error) {
	prefixString := prefix.String()
	if prefixString != "/" {
		prefixString = prefixString + "/"
	}
	prefixAsByteArray := []byte(prefixString)

	iterator.resultsBuilder = dsq.NewResultBuilder(iterator.query)

	iterator.resultsBuilder.Process.Go(func(worker goprocess.Process) {
		iterator.txn.ds.closeLk.RLock()
		iterator.closedEarly = false
		defer func() {
			iterator.txn.ds.closeLk.RUnlock()
			if iterator.closedEarly {
				select {
				case iterator.resultsBuilder.Output <- dsq.Result{
					Error: ErrClosed,
				}:
				case <-iterator.resultsBuilder.Process.Closing():
				}
			}

		}()
		if iterator.txn.ds.closed {
			iterator.closedEarly = true
			return
		}
		iterator.iterator.Seek(prefixAsByteArray)

		iterator.scanThroughToOffset(prefixAsByteArray, worker)
		iterator.yieldResults(prefixAsByteArray, worker)
	})

	go iterator.resultsBuilder.Process.CloseAfterChildren() //nolint

	return iterator.resultsBuilder.Results(), nil
}

func (iterator *BadgerIterator) scanThroughToOffset(prefix []byte, worker goprocess.Process) { //  we might also not need/use this at all
	// skip to the offset
	for _ = 0; iterator.skipped < iterator.query.Offset && iterator.iterator.ValidForPrefix(prefix); iterator.next() {
		// On the happy path, we have no filters and we can go
		// on our way.
		if len(iterator.query.Filters) == 0 {
			iterator.skipped++
			continue
		}

		// On the sad path, we need to apply filters before
		// counting the item as "skipped" as the offset comes
		// _after_ the filter.
		item := iterator.iterator.Item()

		matches := true
		check := func(value []byte) error {
			e := dsq.Entry{
				Key:   string(item.Key()),
				Value: value,
				Size:  int(item.ValueSize()), // this function is basically free
			}

			// Only calculate expirations if we need them.
			if iterator.query.ReturnExpirations {
				e.Expiration = expires(item)
			}
			matches = filter(iterator.query.Filters, e)
			return nil
		}

		// Maybe check with the value, only if we need it.
		var err error
		if iterator.query.KeysOnly {
			err = check(nil)
		} else {
			err = item.Value(check)
		}

		if err != nil {
			select {
			case iterator.resultsBuilder.Output <- dsq.Result{Error: err}:
			case <-iterator.txn.ds.closing: // datastore closing.
				iterator.closedEarly = true
				return
			case <-worker.Closing(): // client told us to close early
				return
			}
		}

		if !matches {
			iterator.skipped++
		}
	}
}

func (iterator *BadgerIterator) yieldResults(prefix []byte, worker goprocess.Process) {
	for _ = 0; iterator.query.Limit <= 0 || iterator.sent < iterator.query.Limit; iterator.next() {
		if !iterator.iterator.ValidForPrefix(prefix) {
			return
		}
		item := iterator.iterator.Item()
		e := dsq.Entry{Key: string(item.Key())}

		// Maybe get the value
		var result dsq.Result
		if !iterator.query.KeysOnly {
			b, err := item.ValueCopy(nil)
			if err != nil {
				result = dsq.Result{Error: err}
			} else {
				e.Value = b
				e.Size = len(b)
				result = dsq.Result{Entry: e}
			}
		} else {
			e.Size = int(item.ValueSize())
			result = dsq.Result{Entry: e}
		}

		if iterator.query.ReturnExpirations {
			result.Expiration = expires(item)
		}

		// Finally, filter it (unless we're dealing with an error).
		if result.Error == nil && filter(iterator.query.Filters, e) {
			continue
		}

		select {
		case iterator.resultsBuilder.Output <- result:
			iterator.sent++
		case <-iterator.txn.ds.closing: // datastore closing.
			iterator.closedEarly = true
			return
		case <-worker.Closing(): // client told us to close early
			return
		}
	}
}
