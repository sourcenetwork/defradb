package db

import (
	"encoding/binary"

	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
)

type sequence struct {
	db  *DB
	key ds.Key
	val uint64
}

func (db *DB) getSequence(key string) (*sequence, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}
	seqKey := ds.NewKey("/seq").ChildString(key)
	seq := &sequence{
		db:  db,
		key: seqKey,
		val: uint64(0),
	}

	_, err := seq.get()
	if err != ds.ErrNotFound {
		seq.update()
	} else if err != nil {
		return nil, err
	}

	return seq, nil
}

func (seq *sequence) get() (uint64, error) {
	val, err := seq.db.systemstore.Get(seq.key)
	if err != nil {
		return 0, err
	}
	num := binary.BigEndian.Uint64(val)
	seq.val = num
	return seq.val, nil
}

func (seq *sequence) update() error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], seq.val)
	if err := seq.db.systemstore.Put(seq.key, buf[:]); err != nil {
		return err
	}

	return nil
}

func (seq *sequence) next() (uint64, error) {
	val, err := seq.get()
	if err != nil {
		return 0, err
	}

	seq.val++
	return seq.val, seq.update()
}
