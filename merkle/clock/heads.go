package clock

import (
	"bytes"
	"encoding/binary"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcenetwork/defradb/store"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	logging "github.com/ipfs/go-log"
)

// heads manages the current Merkle-CRDT heads.
type heads struct {
	store     store.DSReaderWriter
	namespace ds.Key
	logger    logging.StandardLogger
}

func newHeads(store store.DSReaderWriter, namespace ds.Key, logger logging.StandardLogger) *heads {
	return &heads{
		store:     store,
		namespace: namespace,
		logger:    logger,
	}
}

func (hh *heads) key(c cid.Cid) ds.Key {
	// /<namespace>/<cid>
	return hh.namespace.Child(dshelp.CidToDsKey(c))
}

func (hh *heads) load(c cid.Cid) (uint64, error) {
	v, err := hh.store.Get(hh.key(c))
	if err != nil {
		return 0, err
	}
	height, n := binary.Uvarint(v)
	if n <= 0 {
		return 0, errors.New("error decoding height")
	}
	return height, nil
}

func (hh *heads) write(store ds.Write, c cid.Cid, height uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, height)
	if n == 0 {
		return errors.New("error encoding height")
	}
	return store.Put(hh.key(c), buf[0:n])
}

func (hh *heads) delete(store ds.Write, c cid.Cid) error {
	err := store.Delete(hh.key(c))
	if err == ds.ErrNotFound {
		return nil
	}
	return err
}

// IsHead returns if a given cid is among the current heads.
func (hh *heads) IsHead(c cid.Cid) (bool, uint64, error) {
	height, err := hh.load(c)
	if err == ds.ErrNotFound {
		return false, 0, nil
	}
	return err == nil, height, err
}

func (hh *heads) Len() (int, error) {
	list, _, err := hh.List()
	return len(list), err
}

// Replace replaces a head with a new cid.
func (hh *heads) Replace(h, c cid.Cid, height uint64) error {
	hh.logger.Infof("replacing DAG head: %s -> %s (new height: %d)", h, c, height)
	var store ds.Write = hh.store
	var err error

	batchingDs, batching := store.(ds.Batching)
	if batching {
		store, err = batchingDs.Batch()
		if err != nil {
			return err
		}
	}

	err = hh.delete(store, h)
	if err != nil {
		return err
	}

	err = hh.write(store, c, height)
	if err != nil {
		return err
	}

	if batching {
		err := store.(ds.Batch).Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (hh *heads) Add(c cid.Cid, height uint64) error {
	hh.logger.Infof("adding new DAG head: %s (height: %d)", c, height)
	return hh.write(hh.store, c, height)
}

// List returns the list of current heads plus the max height.
// @todo Document Heads.List function
func (hh *heads) List() ([]cid.Cid, uint64, error) {
	q := query.Query{
		Prefix:   hh.namespace.String(),
		KeysOnly: false,
	}

	results, err := hh.store.Query(q)
	if err != nil {
		return nil, 0, err
	}
	defer results.Close()

	heads := make([]cid.Cid, 0)
	var maxHeight uint64
	for r := range results.Next() {
		if r.Error != nil {
			return nil, 0, errors.Wrap(r.Error, "Failed to get next query result")
		}
		// fmt.Println(r.Key, hh.namespace.String())
		headKey := ds.NewKey(strings.TrimPrefix(r.Key, hh.namespace.String()))
		headCid, err := dshelp.DsKeyToCid(headKey)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Failed to get CID from key")
		}
		height, n := binary.Uvarint(r.Value)
		if n <= 0 {
			return nil, 0, errors.New("error decocding height")
		}
		heads = append(heads, headCid)
		if height > maxHeight {
			maxHeight = height
		}
	}
	sort.Slice(heads, func(i, j int) bool {
		ci := heads[i].Bytes()
		cj := heads[j].Bytes()
		return bytes.Compare(ci, cj) < 0
	})

	return heads, maxHeight, nil
}
