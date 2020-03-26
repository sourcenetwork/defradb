package crdt

import (
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"
)

var (
	factoryTestLog = logging.Logger("defradb.tests.factory")
)

func newStores() (ds.Datastore, ds.Datastore, *store.DAGStore) {
	root := ds.NewMapDatastore()
	data := namespace.Wrap(root, ds.NewKey("/test/db/data"))
	heads := namespace.Wrap(root, ds.NewKey("/test/db/heads"))
	s := store.NewDAGStore(namespace.Wrap(root, ds.NewKey("/test/db/blocks")))
	return data, heads, s
}

func TestNewBlankFactory(t *testing.T) {
	f := NewFactory(nil, nil, nil)
	if f == nil {
		t.Error("Returned factory is a nil pointer")
	}
}

func TestNewFactoryWithStores(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s)
	if f == nil {
		t.Error("Returned factory is a nil pointer")
	}

	assert.Equal(t, d, f.datastore)
	assert.Equal(t, h, f.headstore)
	assert.Equal(t, s, f.dagstore)
}

func TestFactoryMultiStoreInterface(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s)
	if f == nil {
		t.Error("Returned factory is a nil pointer")
	}

	// check interface implement
	var _ core.MultiStore = f
	// ms = f

	// check interface functions
	assert.Equal(t, d, f.Data())
	assert.Equal(t, h, f.Head())
	assert.Equal(t, s, f.Dag())
	assert.Equal(t, nil, f.Log())
}

func TestFactorySetStores(t *testing.T) {
	f := NewFactory(nil, nil, nil)
	d, h, s := newStores()
	err := f.SetStores(d, h, s)
	assert.Nil(t, err)

	assert.Equal(t, d, f.Data())
	assert.Equal(t, h, f.Head())
	assert.Equal(t, s, f.Dag())
	assert.Equal(t, nil, f.Log())
}

func TestFactoryWithStores(t *testing.T) {
	f := NewFactory(nil, nil, nil)
	d, h, s := newStores()
	f2 := f.WithStores(d, h, s)
	// assert.NotEmpty

	assert.Nil(t, f.Data())
	assert.Nil(t, f.Head())
	assert.Nil(t, f.Dag())
	assert.Nil(t, f.Log())

	assert.Equal(t, d, f2.Data())
	assert.Equal(t, h, f2.Head())
	assert.Equal(t, s, f2.Dag())
	assert.Equal(t, nil, f2.Log())
}

func TestFullFactoryRegister(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s)
	err := f.Register(LWW_REGISTER, &lwwFactoryFn)
	assert.Nil(t, err)
	assert.Equal(t, &lwwFactoryFn, f.crdts[LWW_REGISTER])
}

func TestBlankFactoryRegister(t *testing.T) {
	f := NewFactory(nil, nil, nil)
	err := f.Register(LWW_REGISTER, &lwwFactoryFn)
	assert.Nil(t, err)
	assert.Equal(t, &lwwFactoryFn, f.crdts[LWW_REGISTER])
}

func TestWithStoresFactoryRegister(t *testing.T) {
	f := NewFactory(nil, nil, nil)
	f.Register(LWW_REGISTER, &lwwFactoryFn)
	d, h, s := newStores()
	f2 := f.WithStores(d, h, s)

	assert.Equal(t, &lwwFactoryFn, f2.crdts[LWW_REGISTER])
}

func TestDefaultFactory(t *testing.T) {
	assert.NotNil(t, DefaultFactory)
	assert.Equal(t, &lwwFactoryFn, DefaultFactory.crdts[LWW_REGISTER])
}

func TestFactoryInstanceMissing(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s)

	_, err := f.Instance(LWW_REGISTER, ds.NewKey("MyKey"))
	assert.Equal(t, err, ErrFactoryTypeNoExist)
}

func TestBlankFactoryInstance(t *testing.T) {
	d, h, s := newStores()
	f1 := NewFactory(nil, nil, nil)
	f1.Register(LWW_REGISTER, &lwwFactoryFn)
	f := f1.WithStores(d, h, s)

	crdt, err := f.Instance(LWW_REGISTER, ds.NewKey("MyKey"))
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)
}

func TestFullFactoryInstance(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s)
	f.Register(LWW_REGISTER, &lwwFactoryFn)

	crdt, err := f.Instance(LWW_REGISTER, ds.NewKey("MyKey"))
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)
}

func TestLWWRegisterFactoryFn(t *testing.T) {
	d, h, s := newStores()
	f := NewFactory(d, h, s) // here factory is only needed to satisfy core.MultiStore interface
	f.SetLogger(factoryTestLog)
	crdt := lwwFactoryFn(f)(ds.NewKey("MyKey"))

	lwwreg, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)

	err := lwwreg.Set([]byte("hi"))
	assert.NoError(t, err)
}
