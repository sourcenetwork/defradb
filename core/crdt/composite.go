package crdt

import (
	"bytes"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ugorji/go/codec"
)

type CompositeDAGDelta struct {
	Priority uint64
	Data     []byte
	SubDags  map[string]cid.Cid
}

// GetPriority gets the current priority for this delta
func (delta *CompositeDAGDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta
func (delta *CompositeDAGDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

func (delta *CompositeDAGDelta) Marshal() ([]byte, error) {
	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err := enc.Encode(struct {
		Priority uint64
		Data     []byte
	}{delta.Priority, delta.Data})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (delta *CompositeDAGDelta) Links() map[string]*ipld.Link {
	links := make(map[string]*ipld.Link)
	for path, c := range delta.SubDags {
		links[path] = &ipld.Link{
			Cid: c,
		}
	}

	return links
}
