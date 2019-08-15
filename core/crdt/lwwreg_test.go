package crdt

import (
	"fmt"
	"reflect"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
	"github.com/ugorji/go/codec"
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

func TestLWWRegisterDeltaInit(t *testing.T) {
	delta := &LWWRegDelta{
		data: []byte("test"),
	}

	var _ core.Delta = delta // checks if LWWRegDelta implements core.Delta (also checked in the implementation code, but w.e)
}

func TestLWWRegosterDeltaGetPriority(t *testing.T) {
	delta := &LWWRegDelta{
		data:     []byte("test"),
		priority: uint64(10),
	}

	if delta.GetPriority() != uint64(10) {
		t.Errorf("LWWRegDelta: GetPriority returned incorrect value, want %v, have %v", uint64(10), delta.GetPriority())
	}
}

func TestLWWRegisterDeltaSetPriority(t *testing.T) {
	delta := &LWWRegDelta{
		data: []byte("test"),
	}

	delta.SetPriority(10)

	if delta.GetPriority() != uint64(10) {
		t.Errorf("LWWRegDelta: SetPriority incorrect value, want %v, have %v", uint64(10), delta.GetPriority())
	}
}

func TestLWWRegisterDeltaMarshal(t *testing.T) {
	delta := &LWWRegDelta{
		data:     []byte("test"),
		priority: uint64(10),
	}
	bytes, err := delta.Marshal()
	if err != nil {
		t.Errorf("Marshal returned an error: %v", err)
		return
	}
	if len(bytes) == 0 {
		t.Error("Expected Marhsal to return serialized bytes, but output is empty")
		return
	}

	fmt.Println(bytes)

	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(bytes, h)
	unmarshaledDelta := &LWWRegDelta{}
	err = dec.Decode(unmarshaledDelta)
	if err != nil {
		t.Errorf("Decode returned an error: %v", err)
		return
	}

	if !reflect.DeepEqual(delta.data, unmarshaledDelta.data) {
		t.Errorf("Unmarhsalled data value doesn't match expected. Want %v, have %v", []byte("test"), unmarshaledDelta.data)
		return
	}

	if delta.priority != unmarshaledDelta.priority {
		t.Errorf("Unmarshalled priority value doesn't match. Want %v, have %v", uint64(10), unmarshaledDelta.priority)
	}
}
