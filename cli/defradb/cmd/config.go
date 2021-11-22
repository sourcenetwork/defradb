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
	"github.com/sourcenetwork/defradb/db"
)

type Config struct {
	Database db.Options
}

type DatabaseConfig struct {
	URL     string
	storage string
}

var (
	defaultConfig = Config{
		Database: db.Options{
			Address: "localhost:9181",
			Store:   "badger",
			Badger: db.BadgerOptions{
				Path: "$HOME/.defradb/data",
			},
		},
	}
)
