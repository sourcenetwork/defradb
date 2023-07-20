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
	Filepath string `json:"filepath"`
	// only JSON is supported at the moment
	Format string `json:"format"`
	// pretty print JSON
	Pretty      bool     `json:"pretty"`
	Collections []string `json:"collections"`
}
