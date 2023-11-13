// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package badger

// This is quite similar in principle to:
//  `https://github.com/MikkelHJuul/bIter/blob/main/iterator.go`
//  that John linked - maybe just use/wrap that
import (
	"context"
	"sync"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	goprocess "github.com/jbenet/goprocess"
	badger "github.com/sourcenetwork/badger/v4"

	"github.com/sourcenetwork/defradb/datastore/iterable"
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
	reversedOrder  bool
}

func (t *txn) GetIterator(q dsq.Query) (iterable.Iterator, error) {
	opt := badger.DefaultIteratorOptions
	// Prefetching prevents the re-use of the iterator
	opt.PrefetchValues = false

	var reversedOrder bool
	// Handle ordering
	if len(q.Orders) > 0 {
		switch orderType := q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
			// We order by key by default.
			reversedOrder = false
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// Reverse order by key
			opt.Reverse = true
			reversedOrder = true
		default:
			// to avoid circlar dependencies in tests we define the error locally.
			return nil, ErrOrderType(orderType)
		}
	}

	badgerIterator := t.txn.NewIterator(opt)

	iterator := BadgerIterator{
		iterator:      badgerIterator,
		txn:           *t,
		reversedOrder: reversedOrder,
	}

	return &iterator, nil
}

func (iterator *BadgerIterator) Close() error {
	// There is a race condition between `iterator.iterator.Next()`
	//  and `iterator.iterator.Close()` which we have to protect against here
	iterator.iteratorLock.Lock()
	iterator.iterator.Close()
	iterator.iteratorLock.Unlock()
	return nil
}

func (iterator *BadgerIterator) next() {
	// There is a race condition between `iterator.iterator.Next()`
	//  and `iterator.iterator.Close()` which we have to protect against here
	iterator.iteratorLock.RLock()
	iterator.iterator.Next()
	iterator.iteratorLock.RUnlock()
}

func (iterator *BadgerIterator) IteratePrefix(
	ctx context.Context,
	startPrefix ds.Key,
	endPrefix ds.Key,
) (dsq.Results, error) {
	formattedStartPrefix := startPrefix.String()
	formattedEndPrefix := endPrefix.String()

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
		iterator.iterator.Seek([]byte(formattedStartPrefix))

		iterator.scanThroughToOffset(formattedStartPrefix, formattedEndPrefix, worker)
		iterator.yieldResults(formattedStartPrefix, formattedEndPrefix, worker)
	})

	go iterator.resultsBuilder.Process.CloseAfterChildren() //nolint:errcheck

	return iterator.resultsBuilder.Results(), nil
}

type itemKeyValidator = func(key string, startPrefix string, endPrefix string) bool

func isValidAscending(key string, startPrefix string, endPrefix string) bool {
	return key >= startPrefix && key <= endPrefix
}

func isValidDescending(key string, startPrefix string, endPrefix string) bool {
	return key <= startPrefix && key >= endPrefix
}

func (iterator *BadgerIterator) getItemKeyValidator() itemKeyValidator {
	if iterator.reversedOrder {
		return isValidDescending
	}
	return isValidAscending
}

func (iterator *BadgerIterator) scanThroughToOffset(
	startPrefix string,
	endPrefix string,
	worker goprocess.Process,
) { //  we might also not need/use this at all
	itemKeyValidator := iterator.getItemKeyValidator()

	// skip to the offset
	for _ = 0; iterator.skipped < iterator.query.Offset &&
		iterator.iterator.Valid(); iterator.next() {
		item := iterator.iterator.Item()
		key := string(item.Key())
		if !itemKeyValidator(key, startPrefix, endPrefix) {
			return
		}

		// On the happy path, we have no filters and we can go
		// on our way.
		if len(iterator.query.Filters) == 0 {
			iterator.skipped++
			continue
		}

		matches := true
		check := func(value []byte) error {
			e := dsq.Entry{
				Key:   key,
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

func (iterator *BadgerIterator) yieldResults(
	startPrefix string,
	endPrefix string,
	worker goprocess.Process,
) {
	itemKeyValidator := iterator.getItemKeyValidator()

	for _ = 0; iterator.query.Limit <= 0 || iterator.sent < iterator.query.Limit; iterator.next() {
		if !iterator.iterator.Valid() {
			return
		}
		item := iterator.iterator.Item()
		key := string(item.Key())
		if !itemKeyValidator(key, startPrefix, endPrefix) {
			return
		}
		e := dsq.Entry{Key: key}

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
