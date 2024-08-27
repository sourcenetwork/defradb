// Code copy-pasted from https://github.com/ipfs/go-ds-badger2/blob/master/datastore.go
// then badger version updated to version 3, and some non-compiling badger defaults
// removed from `init()`

package badger

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	logger "github.com/ipfs/go-log/v2"
	goprocess "github.com/jbenet/goprocess"
	badger "github.com/sourcenetwork/badger/v4"
	"go.uber.org/zap"

	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/errors"
)

var log = logger.Logger("badger")

type Datastore struct {
	DB *badger.DB

	closeLk   sync.RWMutex
	closed    bool
	closeOnce sync.Once
	closing   chan struct{}

	gcDiscardRatio float64
	gcSleep        time.Duration
	gcInterval     time.Duration

	syncWrites bool
}

// Implements the datastore.Batch interface, enabling batching support for
// the badger Datastore.
type batch struct {
	ds         *Datastore
	writeBatch *badger.WriteBatch
}

// Implements the datastore.Txn interface, enabling transaction support for
// the badger Datastore.
type txn struct {
	ds  *Datastore
	txn *badger.Txn

	// Whether this transaction has been implicitly created as a result of a direct Datastore
	// method invocation.
	implicit bool
}

// Options are the badger datastore options, reexported here for convenience.
type Options struct {
	// Please refer to the Badger docs to see what this is for
	GcDiscardRatio float64

	// Interval between GC cycles
	//
	// If zero, the datastore will perform no automatic garbage collection.
	GcInterval time.Duration

	// Sleep time between rounds of a single GC cycle.
	//
	// If zero, the datastore will only perform one round of GC per
	// GcInterval.
	GcSleep time.Duration

	badger.Options
}

// DefaultOptions are the default options for the badger datastore.
var DefaultOptions Options

func init() {
	DefaultOptions = Options{
		GcDiscardRatio: 0.2,
		GcInterval:     15 * time.Minute,
		GcSleep:        10 * time.Second,
		Options:        badger.DefaultOptions(""),
	}
	// This is to optimize the database on close so it can be opened
	// read-only and efficiently queried. We don't do that and hanging on
	// stop isn't nice.
	DefaultOptions.Options.CompactL0OnClose = false
	/*
		// The alternative is "crash on start and tell the user to fix it". This
		// will truncate corrupt and unsynced data, which we don't guarantee to
		// persist anyways.
		DefaultOptions.Options.Truncate = true

		// Uses less memory, is no slower when writing, and is faster when
		// reading (in some tests).
		DefaultOptions.Options.ValueLogLoadingMode = options.FileIO

		// Explicitly set this to mmap. This doesn't use much memory anyways.
		DefaultOptions.Options.TableLoadingMode = options.MemoryMap

		// Reduce this from 64MiB to 16MiB. That means badger will hold on to
		// 20MiB by default instead of 80MiB.
		//
		// This does not appear to have a significant performance hit.
		DefaultOptions.Options.MaxTableSize = 16 << 20*/
}

var _ ds.Datastore = (*Datastore)(nil)
var _ ds.TxnDatastore = (*Datastore)(nil)
var _ ds.GCDatastore = (*Datastore)(nil)
var _ ds.Batching = (*Datastore)(nil)

// NewDatastore creates a new badger datastore.
//
// DO NOT set the Dir and/or ValuePath fields of opt, they will be set for you.
func NewDatastore(path string, options *Options) (*Datastore, error) {
	// Copy the options because we modify them.
	var opt badger.Options
	var gcDiscardRatio float64
	var gcSleep time.Duration
	var gcInterval time.Duration
	if options == nil {
		opt = DefaultOptions.Options
		gcDiscardRatio = DefaultOptions.GcDiscardRatio
		gcSleep = DefaultOptions.GcSleep
		gcInterval = DefaultOptions.GcInterval
	} else {
		opt = options.Options
		gcDiscardRatio = options.GcDiscardRatio
		gcSleep = options.GcSleep
		gcInterval = options.GcInterval
	}

	if gcSleep <= 0 {
		// If gcSleep is 0, we don't perform multiple rounds of GC per
		// cycle.
		gcSleep = gcInterval
	}

	if !opt.InMemory {
	    opt.Dir = path
	    opt.ValueDir = path
    }

	opt.Logger = &compatLogger{
		SugaredLogger: *log.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		skipLogger:    *log.Desugar().WithOptions(zap.AddCallerSkip(2)).Sugar(),
	}

	kv, err := badger.Open(opt)
	if err != nil {
		if strings.HasPrefix(err.Error(), "manifest has unsupported version:") {
			err = errors.Wrap("unsupported badger version, use github.com/ipfs/badgerds-upgrade to upgrade", err)
		}
		return nil, err
	}

	ds := &Datastore{
		DB:             kv,
		closing:        make(chan struct{}),
		gcDiscardRatio: gcDiscardRatio,
		gcSleep:        gcSleep,
		gcInterval:     gcInterval,
		syncWrites:     opt.SyncWrites,
	}

	// Start the GC process if requested.
	if ds.gcInterval > 0 {
		go ds.periodicGC()
	}

	return ds, nil
}

// Keep scheduling GC's AFTER `gcInterval` has passed since the previous GC
func (d *Datastore) periodicGC() {
	gcTimeout := time.NewTimer(d.gcInterval)
	defer gcTimeout.Stop()

	for {
		select {
		case <-gcTimeout.C:
			err := d.gcOnce()
			switch {
			case errors.Is(err, badger.ErrNoRewrite) || errors.Is(err, badger.ErrRejected):
				// No rewrite means we've fully garbage collected.
				// Rejected means someone else is running a GC
				// or we're closing.
				gcTimeout.Reset(d.gcInterval)
			case err == nil:
				gcTimeout.Reset(d.gcSleep)
			case errors.Is(err, ErrClosed):
				return
			default:
				log.Errorf("Error during a GC cycle: %s", err)
				// Not much we can do on a random error but log it and continue.
				gcTimeout.Reset(d.gcInterval)
			}
		case <-d.closing:
			return
		}
	}
}

// NewTransaction starts a new transaction. The resulting transaction object
// can be mutated without incurring changes to the underlying Datastore until
// the transaction is Committed.
func (d *Datastore) NewTransaction(ctx context.Context, readOnly bool) (ds.Txn, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}

	return &txn{d, d.DB.NewTransaction(!readOnly), false}, nil
}

// newImplicitTransaction creates a transaction marked as 'implicit'.
// Implicit transactions are created by Datastore methods performing single operations.
func (d *Datastore) newImplicitTransaction(readOnly bool) *txn {
	return &txn{d, d.DB.NewTransaction(!readOnly), true}
}

func (d *Datastore) NewIterableTransaction(
	ctx context.Context,
	readOnly bool,
) (iterable.IterableTxn, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}

	return &txn{d, d.DB.NewTransaction(!readOnly), false}, nil
}

func (d *Datastore) Put(ctx context.Context, key ds.Key, value []byte) error {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}

	txn := d.newImplicitTransaction(false)
	defer txn.discard()

	if err := txn.put(key, value); err != nil {
		return err
	}

	return txn.commit()
}

func (d *Datastore) Sync(ctx context.Context, prefix ds.Key) error {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}

	if d.syncWrites {
		return nil
	}

	return d.DB.Sync()
}

func (d *Datastore) Get(ctx context.Context, key ds.Key) (value []byte, err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}

	txn := d.newImplicitTransaction(true)
	defer txn.discard()

	return txn.get(key)
}

func (d *Datastore) Has(ctx context.Context, key ds.Key) (bool, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return false, ErrClosed
	}

	txn := d.newImplicitTransaction(true)
	defer txn.discard()

	return txn.has(key)
}

func (d *Datastore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return -1, ErrClosed
	}

	txn := d.newImplicitTransaction(true)
	defer txn.discard()

	return txn.getSize(key)
}

func (d *Datastore) Delete(ctx context.Context, key ds.Key) error {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()

	txn := d.newImplicitTransaction(false)
	defer txn.discard()

	err := txn.delete(key)
	if err != nil {
		return err
	}

	return txn.commit()
}

func (d *Datastore) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}

	txn := d.newImplicitTransaction(true)
	// We cannot defer txn.Discard() here, as the txn must remain active while the iterator is open.
	// https://github.com/dgraph-io/badger/commit/b1ad1e93e483bbfef123793ceedc9a7e34b09f79
	// The closing logic in the query goprocess takes care of discarding the implicit transaction.
	return txn.query(q)
}

// DiskUsage implements the PersistentDatastore interface.
// It returns the sum of lsm and value log files sizes in bytes.
func (d *Datastore) DiskUsage(ctx context.Context) (uint64, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return 0, ErrClosed
	}
	lsm, vlog := d.DB.Size()
	return uint64(lsm + vlog), nil
}

func (d *Datastore) Close() error {
	d.closeOnce.Do(func() {
		close(d.closing)
	})
	d.closeLk.Lock()
	defer d.closeLk.Unlock()
	if d.closed {
		return ErrClosed
	}
	d.closed = true
	return d.DB.Close()
}

// Batch creates a new Batch object. This provides a way to do many writes, when
// there may be too many to fit into a single transaction.
func (d *Datastore) Batch(ctx context.Context) (ds.Batch, error) {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return nil, ErrClosed
	}

	b := &batch{d, d.DB.NewWriteBatch()}
	// Ensure that incomplete transaction resources are cleaned up in case
	// batch is abandoned.
	runtime.SetFinalizer(b, func(b *batch) {
		b.cancel()
		log.Error("Batch not committed or canceled")
	})

	return b, nil
}

func (d *Datastore) CollectGarbage(ctx context.Context) (err error) {
	// The idea is to keep calling DB.RunValueLogGC() till Badger no longer has any log files
	// to GC(which would be indicated by an error, please refer to Badger GC docs).
	for err == nil {
		err = d.gcOnce()
	}

	if errors.Is(err, badger.ErrNoRewrite) {
		err = nil
	}

	return err
}

func (d *Datastore) gcOnce() error {
	d.closeLk.RLock()
	defer d.closeLk.RUnlock()
	if d.closed {
		return ErrClosed
	}
	return d.DB.RunValueLogGC(d.gcDiscardRatio)
}

var _ ds.Batch = (*batch)(nil)

func (b *batch) Put(ctx context.Context, key ds.Key, value []byte) error {
	b.ds.closeLk.RLock()
	defer b.ds.closeLk.RUnlock()
	if b.ds.closed {
		return ErrClosed
	}
	return b.put(key, value)
}

func (b *batch) put(key ds.Key, value []byte) error {
	return b.writeBatch.Set(key.Bytes(), value)
}

func (b *batch) Delete(ctx context.Context, key ds.Key) error {
	b.ds.closeLk.RLock()
	defer b.ds.closeLk.RUnlock()
	if b.ds.closed {
		return ErrClosed
	}

	return b.delete(key)
}

func (b *batch) delete(key ds.Key) error {
	return b.writeBatch.Delete(key.Bytes())
}

func (b *batch) Commit(ctx context.Context) error {
	b.ds.closeLk.RLock()
	defer b.ds.closeLk.RUnlock()
	if b.ds.closed {
		return ErrClosed
	}

	return b.commit()
}

func (b *batch) commit() error {
	err := b.writeBatch.Flush()
	if err != nil {
		// Discard incomplete transaction held by b.writeBatch
		b.cancel()
		return err
	}
	runtime.SetFinalizer(b, nil)
	return nil
}

func (b *batch) cancel() {
	b.writeBatch.Cancel()
	runtime.SetFinalizer(b, nil)
}

var _ ds.Txn = (*txn)(nil)

func (t *txn) Put(ctx context.Context, key ds.Key, value []byte) error {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return ErrClosed
	}
	return t.put(key, value)
}

func (t *txn) put(key ds.Key, value []byte) error {
	return t.txn.Set(key.Bytes(), value)
}

func (t *txn) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return nil, ErrClosed
	}

	return t.get(key)
}

func (t *txn) get(key ds.Key) ([]byte, error) {
	item, err := t.txn.Get(key.Bytes())
	if errors.Is(err, badger.ErrKeyNotFound) {
		err = ds.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return item.ValueCopy(nil)
}

func (t *txn) Has(ctx context.Context, key ds.Key) (bool, error) {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return false, ErrClosed
	}

	return t.has(key)
}

func (t *txn) has(key ds.Key) (bool, error) {
	_, err := t.txn.Get(key.Bytes())
	switch {
	case errors.Is(err, badger.ErrKeyNotFound):
		return false, nil
	case err == nil:
		return true, nil
	default:
		return false, err
	}
}

func (t *txn) GetSize(ctx context.Context, key ds.Key) (int, error) {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return -1, ErrClosed
	}

	return t.getSize(key)
}

func (t *txn) getSize(key ds.Key) (int, error) {
	item, err := t.txn.Get(key.Bytes())
	switch {
	case err == nil:
		size := int(item.ValueSize())
		if size == 0 {
			val, err := item.ValueCopy(nil)
			if err != nil {
				return 0, err
			}
			size = len(val)
		}
		return size, nil
	case errors.Is(err, badger.ErrKeyNotFound):
		return -1, ds.ErrNotFound
	default:
		return -1, err
	}
}

func (t *txn) Delete(ctx context.Context, key ds.Key) error {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return ErrClosed
	}

	return t.delete(key)
}

func (t *txn) delete(key ds.Key) error {
	return t.txn.Delete(key.Bytes())
}

func (t *txn) Query(ctx context.Context, q dsq.Query) (dsq.Results, error) {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return nil, ErrClosed
	}

	return t.query(q)
}

func (t *txn) query(q dsq.Query) (dsq.Results, error) {
	opt := badger.DefaultIteratorOptions
	opt.PrefetchValues = !q.KeysOnly

	prefix := ds.NewKey(q.Prefix).String()
	if prefix != "/" {
		opt.Prefix = []byte(prefix + "/")
	}

	// Handle ordering
	if len(q.Orders) > 0 {
		switch q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
		// We order by key by default.
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// Reverse order by key
			opt.Reverse = true
		default:
			// Ok, we have a weird order we can't handle. Let's
			// perform the _base_ query (prefix, filter, etc.), then
			// handle sort/offset/limit later.

			// Skip the stuff we can't apply.
			baseQuery := q
			baseQuery.Limit = 0
			baseQuery.Offset = 0
			baseQuery.Orders = nil

			// perform the base query.
			res, err := t.query(baseQuery)
			if err != nil {
				return nil, err
			}

			// fix the query
			res = dsq.ResultsReplaceQuery(res, q)

			// Remove the parts we've already applied.
			naiveQuery := q
			naiveQuery.Prefix = ""
			naiveQuery.Filters = nil

			// Apply the rest of the query
			return dsq.NaiveQueryApply(naiveQuery, res), nil
		}
	}

	it := t.txn.NewIterator(opt)
	qrb := dsq.NewResultBuilder(q)
	qrb.Process.Go(func(worker goprocess.Process) {
		t.ds.closeLk.RLock()
		closedEarly := false
		defer func() {
			t.ds.closeLk.RUnlock()
			if closedEarly {
				select {
				case qrb.Output <- dsq.Result{
					Error: ErrClosed,
				}:
				case <-qrb.Process.Closing():
				}
			}
		}()
		if t.ds.closed {
			closedEarly = true
			return
		}

		// this iterator is part of an implicit transaction, so when
		// we're done we must discard the transaction. It's safe to
		// discard the txn it because it contains the iterator only.
		if t.implicit {
			defer t.discard()
		}

		defer it.Close()

		// All iterators must be started by rewinding.
		it.Rewind()

		// skip to the offset
		for skipped := 0; skipped < q.Offset && it.Valid(); it.Next() {
			// On the happy path, we have no filters and we can go
			// on our way.
			if len(q.Filters) == 0 {
				skipped++
				continue
			}

			// On the sad path, we need to apply filters before
			// counting the item as "skipped" as the offset comes
			// _after_ the filter.
			item := it.Item()

			matches := true
			check := func(value []byte) error {
				e := dsq.Entry{
					Key:   string(item.Key()),
					Value: value,
					Size:  int(item.ValueSize()), // this function is basically free
				}

				// Only calculate expirations if we need them.
				if q.ReturnExpirations {
					e.Expiration = expires(item)
				}
				matches = filter(q.Filters, e)
				return nil
			}

			// Maybe check with the value, only if we need it.
			var err error
			if q.KeysOnly {
				err = check(nil)
			} else {
				err = item.Value(check)
			}

			if err != nil {
				select {
				case qrb.Output <- dsq.Result{Error: err}:
				case <-t.ds.closing: // datastore closing.
					closedEarly = true
					return
				case <-worker.Closing(): // client told us to close early
					return
				}
			}
			if !matches {
				skipped++
			}
		}

		for sent := 0; (q.Limit <= 0 || sent < q.Limit) && it.Valid(); it.Next() {
			item := it.Item()
			e := dsq.Entry{Key: string(item.Key())}

			// Maybe get the value
			var result dsq.Result
			if !q.KeysOnly {
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

			if q.ReturnExpirations {
				result.Expiration = expires(item)
			}

			// Finally, filter it (unless we're dealing with an error).
			if result.Error == nil && filter(q.Filters, e) {
				continue
			}

			select {
			case qrb.Output <- result:
				sent++
			case <-t.ds.closing: // datastore closing.
				closedEarly = true
				return
			case <-worker.Closing(): // client told us to close early
				return
			}
		}
	})

	//nolint:errcheck
	go qrb.Process.CloseAfterChildren()

	return qrb.Results(), nil
}

func (t *txn) Commit(ctx context.Context) error {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return ErrClosed
	}

	return t.commit()
}

func (t *txn) commit() error {
	err := t.txn.Commit()
	return convertError(err)
}

func (t *txn) Discard(ctx context.Context) {
	t.ds.closeLk.RLock()
	defer t.ds.closeLk.RUnlock()
	if t.ds.closed {
		return
	}

	t.discard()
}

func (t *txn) discard() {
	t.txn.Discard()
}

// filter returns _true_ if we should filter (skip) the entry
func filter(filters []dsq.Filter, entry dsq.Entry) bool {
	for _, f := range filters {
		if !f.Filter(entry) {
			return true
		}
	}
	return false
}

func expires(item *badger.Item) time.Time {
	return time.Unix(int64(item.ExpiresAt()), 0)
}
