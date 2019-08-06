package crdt

import (
	"encoding/binary"
	"errors"

	ds "github.com/ipfs/go-datastore"
)

var (
	keysNs         = "k" // /keys namespace /set/k/<key>/{v,p}
	valueSuffix    = "v" // value key
	prioritySuffix = "p" // priority key
)

// baseCRDT is embedded as a base layer into all
// the core CRDT implementations to reduce code
// duplcation, and better manage the overhead
// tasks that all the CRDTs need to implement anyway
type baseCRDT struct {
	store          ds.Datastore
	namespace      ds.Key
	keysNs         string
	valueSuffix    string
	prioritySuffix string
}

// @TODO paramaterize ns/suffix
func newBaseCRDT(store ds.Datastore, namespace ds.Key) baseCRDT {
	return baseCRDT{
		store:          store,
		namespace:      namespace,
		keysNs:         keysNs,
		valueSuffix:    valueSuffix,
		prioritySuffix: prioritySuffix,
	}
}

func (base baseCRDT) keyPrefix(key string) ds.Key {
	return base.namespace.ChildString(key)
}

func (base baseCRDT) valueKey(key string) ds.Key {
	return base.keyPrefix(base.keysNs).ChildString(key).ChildString(base.valueSuffix)
}

func (base baseCRDT) priorityKey(key string) ds.Key {
	return base.keyPrefix(base.keysNs).ChildString(key).ChildString(base.prioritySuffix)
}

func (base baseCRDT) setPriority(key string, priority uint64) error {
	prioK := base.priorityKey(key)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority+1)
	if n == 0 {
		return errors.New("error encoding priority")
	}

	return base.store.Put(prioK, buf[0:n])
}

// get the current priority for given key
func (base baseCRDT) getPriority(key string) (uint64, error) {
	pKey := base.priorityKey(key)
	pbuf, err := base.store.Get(pKey)
	if err != nil {
		if err == ds.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	prio, num := binary.Uvarint(pbuf)
	if num <= 0 {
		return 0, errors.New("failed to decode priority")
	}
	return prio, nil
}
