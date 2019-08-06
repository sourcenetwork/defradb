package crdt

import (
	"reflect"
	"testing"

	ds "github.com/ipfs/go-datastore"
)

func newMockStore() ds.Datastore {
	return ds.NewMapDatastore()
}

func setupLWWRegister() LWWRegister {
	store := newMockStore()
	ns := ds.NewKey("defra/test")
	id := "AAAA-BBBB"
	return NewLWWRegister(store, ns, id)
}

func setupLoadedLWWRegster() LWWRegister {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))
	lww.Merge(addDelta, "test")
	return lww
}

func TestLWWRegisterAddDelta(t *testing.T) {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))

	if !reflect.DeepEqual(addDelta.data, []byte("test")) {
		t.Errorf("Delta unexpected value, was %s want %s", addDelta.data, []byte("test"))
	}
}

func TestLWWRegisterInitialMerge(t *testing.T) {
	lww := setupLWWRegister()
	addDelta := lww.Set([]byte("test"))
	err := lww.Merge(addDelta, "test")
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
		return
	}

	val, err := lww.Value()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	expectedVal := []byte("test")
	if string(val) != string(expectedVal) {
		t.Errorf("Mismatch value for LWWRegister, was %s want %s", val, expectedVal)
	}
}

func TestLWWReisterFollowupMerge(t *testing.T) {
	lww := setupLoadedLWWRegster()
	addDelta := lww.Set([]byte("test2"))
	addDelta.SetPriority(2)
	lww.Merge(addDelta, "test")

	val, err := lww.Value()
	if err != nil {
		t.Error(err)
	}

	if string(val) != string([]byte("test2")) {
		t.Errorf("Incorrect merge state, want %s, have %s", []byte("test2"), val)
	}
}

func TestLWWRegisterOldMerge(t *testing.T) {
	lww := setupLoadedLWWRegster()
	addDelta := lww.Set([]byte("test-1"))
	addDelta.SetPriority(0)
	lww.Merge(addDelta, "test")

	val, err := lww.Value()
	if err != nil {
		t.Error(err)
	}

	if string(val) != string([]byte("test")) {
		t.Errorf("Incorrect merge state, want %s, have %s", []byte("test"), val)
	}
}
