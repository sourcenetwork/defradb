// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package cmd

import (
	badgerds "github.com/ipfs/go-ds-badger"
	"github.com/sourcenetwork/defradb/db"
)

type Config struct {
	Database Options
}

type Options struct {
	Address string
	Store   string
	Memory  MemoryOptions
	Badger  BadgerOptions
}

// BadgerOptions for the badger instance of the backing datastore
type BadgerOptions struct {
	Path string
	*badgerds.Options
}

// MemoryOptions for the memory instance of the backing datastore
type MemoryOptions struct {
	Size uint64
}

func (cmdOptions *Options) ToDatabaseOptions() db.Options {
	return db.Options{
		Store: cmdOptions.Store,
		Memory: db.MemoryOptions{
			Size: cmdOptions.Memory.Size,
		},
		Badger: db.BadgerOptions{
			Path:    cmdOptions.Badger.Path,
			Options: cmdOptions.Badger.Options,
		},
	}
}

var (
	defaultConfig = Config{
		Database: Options{
			Address: "localhost:9181",
			Store:   "badger",
			Badger: BadgerOptions{
				Path: "$HOME/.defradb/data",
			},
		},
	}
)
