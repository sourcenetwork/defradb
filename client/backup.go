// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"
)

// Backup contains DefraDB's supported backup operations.
type Backup interface {
	// BasicImport imports a json dataset.
	// filepath must be accessible to the node.
	BasicImport(ctx context.Context, filepath string) error
	// BasicExport exports the current data or subset of data to file in json format.
	BasicExport(ctx context.Context, config *BackupConfig) error
}

// BackupConfig holds the configuration parameters for database backups.
type BackupConfig struct {
	// If a file already exists at this location, it will be truncated and overwriten.
	Filepath string `json:"filepath"`
	// Only JSON is supported at the moment.
	Format string `json:"format"`
	// List of collection names to select which one to backup.
	Collections []string `json:"collections"`
	// Pretty print JSON.
	Pretty bool `json:"pretty"`
}
