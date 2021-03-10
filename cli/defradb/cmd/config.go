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
	// badger
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
