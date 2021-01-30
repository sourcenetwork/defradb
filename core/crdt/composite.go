package crdt

import (
	"bytes"
	"errors"
	"sort"
	"strings"

	"github.com/sourcenetwork/defradb/core"

	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ugorji/go/codec"
)

var (
	_ core.ReplicatedData = (*CompositeDAG)(nil)
	_ core.CompositeDelta = (*CompositeDAGDelta)(nil)
)

type CompositeDAGDelta struct {
	Priority uint64
	Data     []byte
	SubDAGs  []core.DAGLink
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

func (delta *CompositeDAGDelta) Value() interface{} {
	return delta.Data
}

func (delta *CompositeDAGDelta) Links() []core.DAGLink {
	// links := make(map[string]*ipld.Link)
	// for path, c := range delta.SubDAGs {
	// 	links[path] = &ipld.Link{
	// 		Cid: c,
	// 	}
	// }

	// return links
	return delta.SubDAGs
}

// CompositeDAG is a CRDT structure that is used
// to track a collcetion of sub MerkleCRDTs.
type CompositeDAG struct {
	baseCRDT
	key   string
	data  []byte
	links map[string]cid.Cid
}

func NewCompositeDAG(store core.DSReaderWriter, namespace ds.Key, key string) CompositeDAG {
	return CompositeDAG{}
}

func (c CompositeDAG) Value() ([]byte, error) {
	return nil, nil
}

func (c CompositeDAG) Set(patch []byte, links []core.DAGLink) *CompositeDAGDelta {
	// make sure the links are sorted lexigraphically by CID
	sort.Slice(links, func(i, j int) bool {
		return strings.Compare(links[i].Cid.String(), links[j].Cid.String()) < 0
	})
	return &CompositeDAGDelta{
		Data:    patch,
		SubDAGs: links,
	}
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
// @todo
func (c CompositeDAG) Merge(delta core.Delta, id string) error {
	// d, ok := delta.(*CompositeDAGDelta)
	// if !ok {
	// 	return core.ErrMismatchedMergeType
	// }

	// return reg.setValue(d.Data, d.GetPriority())
	return nil
}

// DeltaDecode is a typed helper to extract
// a LWWRegDelta from a ipld.Node
// for now lets do cbor (quick to implement)
func (c CompositeDAG) DeltaDecode(node ipld.Node) (core.Delta, error) {
	delta := &CompositeDAGDelta{}
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, errors.New("Failed to cast ipld.Node to ProtoNode")
	}
	data := pbNode.Data()
	// fmt.Println(data)
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(data, h)
	err := dec.Decode(delta)
	if err != nil {
		return nil, err
	}

	// get links
	for _, link := range pbNode.Links() {
		if link.Name == "head" { // ignore the head links
			continue
		}

		delta.SubDAGs = append(delta.SubDAGs, core.DAGLink{
			Name: link.Name,
			Cid:  link.Cid,
		})
	}
	return delta, nil
}
