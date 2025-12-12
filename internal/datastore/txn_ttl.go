package datastore

import (
	"context"
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/internal/ttl"
	"github.com/sourcenetwork/immutable"
)

type BasicTTLTxn struct {
	*BasicTxn
	cache *ttl.Cache[uint64, Txn]
}

// NewTTLTxnFrom creates a new transaction bound by a idle ttl, meaning if theres
// no operations on the transaction for a duration equal to the expiration ttl, it
// will automatically discard.
//
// The actual automatic transaction discard functionality must be implemented on
// the `ttl.Wheel` or `ttl.Cache` `onExpire` funtion. The `BasicTTLTxn` is only
// responsible for ensuring the ttl is correctly updated.
func NewTTLTxnFrom(
	rootstore corekv.TxnStore,
	cache *ttl.Cache[uint64, Txn],
	id uint64,
	readonly bool,
	updateTTL time.Duration,
	chunkSize immutable.Option[int],
) (*BasicTTLTxn, error) {
	rootTxn := rootstore.NewTxn(readonly)
	rootTTLTxn, err := ttlWrapStore(rootTxn, cache, id, updateTTL)
	if err != nil {
		return nil, err
	}

	multistore := NewMultistore(rootTTLTxn, chunkSize)
	return &BasicTTLTxn{
		BasicTxn: &BasicTxn{
			Multistore: multistore,
			txn:        rootTxn,
			id:         id,
		},
		cache: cache,
	}, nil
}

func (t *BasicTTLTxn) Commit() error {
	// regardless of the error result of txn.Commit()
	// we need to delete the entry from the wheel
	// since transactions are considered invalid if
	// Commit() returns any error
	defer t.cache.Delete(t.ID())
	return t.BasicTxn.Commit()
}

func (t *BasicTTLTxn) Discard() {
	t.BasicTxn.Discard()
	t.cache.Delete(t.ID())
}

// ttlReaderWriter wraps a given corekv.ReaderWriter
// and for each op will refresh the timingwheel with
// the specified ttl.
type ttlReaderWriter struct {
	root      corekv.ReaderWriter
	updateTTL time.Duration
	cache     *ttl.Cache[uint64, Txn]
	cacheKey  uint64
}

func ttlWrapStore(store corekv.ReaderWriter, cache *ttl.Cache[uint64, Txn], cacheKey uint64, updateTTL time.Duration) (*ttlReaderWriter, error) {
	return &ttlReaderWriter{
		root:      store,
		updateTTL: updateTTL,
		cache:     cache,
		cacheKey:  cacheKey,
	}, nil
}

// Get returns the value at the given key.
// If no item with the given key is found, nil, and an [ErrNotFound]
// error will be returned.
func (rw *ttlReaderWriter) Get(ctx context.Context, key []byte) ([]byte, error) {
	err := rw.cache.UpdateTTL(rw.cacheKey, rw.updateTTL)
	if err != nil {
		return nil, err
	}

	return rw.root.Get(ctx, key)
}

// Has returns true if an item at the given key is found, otherwise
// will return false.
func (rw *ttlReaderWriter) Has(ctx context.Context, key []byte) (bool, error) {
	err := rw.cache.UpdateTTL(rw.cacheKey, rw.updateTTL)
	if err != nil {
		return false, err
	}

	return rw.root.Has(ctx, key)
}

// Iterator returns a read-only iterator using the given options.
func (rw *ttlReaderWriter) Iterator(ctx context.Context, opts corekv.IterOptions) (corekv.Iterator, error) {
	err := rw.cache.UpdateTTL(rw.cacheKey, rw.updateTTL)
	if err != nil {
		return nil, err
	}

	return rw.root.Iterator(ctx, opts)
}

// Set sets the value stored against the given key.
//
// If an item already exists at the given key it will be overwritten.
func (rw *ttlReaderWriter) Set(ctx context.Context, key []byte, value []byte) error {
	err := rw.cache.UpdateTTL(rw.cacheKey, rw.updateTTL)
	if err != nil {
		return err
	}

	return rw.root.Set(ctx, key, value)
}

// Delete removes the value at the given key.
//
// If no matching key is found the behaviour is undefined:
// https://github.com/sourcenetwork/corekv/issues/36
func (rw *ttlReaderWriter) Delete(ctx context.Context, key []byte) error {
	err := rw.cache.UpdateTTL(rw.cacheKey, rw.updateTTL)
	if err != nil {
		return err
	}

	return rw.root.Delete(ctx, key)
}
