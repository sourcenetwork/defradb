// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"bytes"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"

	"github.com/sourcenetwork/defradb/errors"
)

var (
	// Individual Store Keys
	systemStoreKey = byte('s')
	dataStoreKey   = byte('d')
	headStoreKey   = byte('h')
	blockStoreKey  = byte('b')
	peerStoreKey   = byte('p')
	encStoreKey    = byte('e')
)

type Multistore struct {
	block  Blockstore
	data   corekv.ReaderWriter
	enc    Blockstore
	head   corekv.ReaderWriter
	peer   corekv.ReaderWriter
	root   corekv.ReaderWriter
	system corekv.ReaderWriter
}

func NewMultistore(rootstore corekv.ReaderWriter) *Multistore {
	return &Multistore{
		block:  newBlockstore(namespace.Wrap(rootstore, []byte{blockStoreKey})),
		data:   namespace.Wrap(rootstore, []byte{dataStoreKey}),
		enc:    newBlockstore(namespace.Wrap(rootstore, []byte{encStoreKey})),
		head:   namespace.Wrap(rootstore, []byte{headStoreKey}),
		peer:   namespace.Wrap(rootstore, []byte{peerStoreKey}),
		root:   rootstore,
		system: namespace.Wrap(rootstore, []byte{systemStoreKey}),
	}
}

func (m *Multistore) Blockstore() Blockstore {
	return m.block
}

func (m *Multistore) Datastore() corekv.ReaderWriter {
	return m.data
}

func (m *Multistore) Encstore() Blockstore {
	return m.enc
}

func (m *Multistore) Headstore() corekv.ReaderWriter {
	return m.head
}

func (m *Multistore) Peerstore() corekv.ReaderWriter {
	return m.peer
}

func (m *Multistore) Rootstore() corekv.ReaderWriter {
	return m.root
}

func (m *Multistore) Systemstore() corekv.ReaderWriter {
	return m.system
}

func DatastoreFrom(rootstore corekv.ReaderWriter) corekv.ReaderWriter {
	return namespace.Wrap(rootstore, []byte{dataStoreKey})
}

func EncstoreFrom(rootstore corekv.ReaderWriter) Blockstore {
	return newBlockstore(namespace.Wrap(rootstore, []byte{encStoreKey}))
}

func HeadstoreFrom(rootstore corekv.ReaderWriter) corekv.ReaderWriter {
	return namespace.Wrap(rootstore, []byte{headStoreKey})
}

func BlockstoreFrom(rootstore corekv.ReaderWriter) Blockstore {
	return newBlockstore(namespace.Wrap(rootstore, []byte{blockStoreKey}))
}

func P2PBlockstoreFrom(rootstore corekv.ReaderWriter) Blockstore {
	return &p2pBlockStore{
		bstore: newBlockstore(namespace.Wrap(rootstore, []byte{blockStoreKey})),
	}
}

func SystemstoreFrom(rootstore corekv.ReaderWriter) corekv.ReaderWriter {
	return namespace.Wrap(rootstore, []byte{systemStoreKey})
}

func PeerstoreFrom(rootstore corekv.ReaderWriter) corekv.ReaderWriter {
	return namespace.Wrap(rootstore, []byte{peerStoreKey})
}

// HumanReadableKey converts a raw byte and representation of a key into a human redable format.
func HumanReadableKey(key []byte) (string, error) {
	switch key[0] {
	case blockStoreKey:
		if bytes.HasPrefix(key[1:], []byte{toMergeIndexPrefix}) {
			cid, err := cid.Cast(key[2:])
			if err != nil {
				return "", errors.WithStack(err)
			}
			return "blocks/to_merge/" + cid.String(), nil
		}
		cid, err := cid.Cast(key[1:])
		if err != nil {
			return "", errors.WithStack(err)
		}
		return "blocks/" + cid.String(), nil
	case dataStoreKey:
		return "data" + string(key[1:]), nil
	case encStoreKey:
		cid, err := cid.Cast(key[1:])
		if err != nil {
			return "", errors.WithStack(err)
		}
		return "encryption/" + cid.String(), nil
	case encStoreKey:
		return "system" + string(key[1:]), nil
	case headStoreKey:
		return "heads" + string(key[1:]), nil
	case peerStoreKey:
		return "peers" + string(key[1:]), nil
	case systemStoreKey:
		return "system" + string(key[1:]), nil
	}
	return string(key), nil
}
