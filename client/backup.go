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

// BackupConfig holds the configuration parameters for database backups.
type BackupConfig struct {
	// If a file already exists at this location, it will be truncated and overwriten.
	Filepath string `json:"filepath"`
	// Only JSON is supported at the moment.
	Format string `json:"format"`
	// Pretty print JSON.
	Pretty bool `json:"pretty"`
	// List of collection names to select which one to backup.
	Collections []string `json:"collections"`
}
