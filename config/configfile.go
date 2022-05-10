// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package config

import (
	"fmt"
	"os"
)

const (
	defaultDirPerm        = 0o700
	defaultConfigFilePerm = 0o644
)

func (cfg *Config) writeConfigFile(path string) error {
	buffer, err := cfg.toBytes()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, buffer, defaultConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (cfg *Config) WriteConfigFileToRootDir(rootDir string) error {
	path := fmt.Sprintf("%v/%v", rootDir, defaultDefraDBConfigFileName)
	return cfg.writeConfigFile(path)
}

// Must reflect Config in content and configuration.
// All parameters must be represented here, to support Viper's automatic environment variable handling.
const defaultConfigTemplate = `# DefraDB configuration (YAML)

# NOTE: Paths below are relative to the DefraDB root directory.
# By default, the DefraDB root directory is "$HOME/.defradb", but
# can be changed via the $DEFRA_ROOT env variable or --rootdir CLI flag.

datastore:
  # Store can be badger | memory
    #   badger: fast pure Go key-value store optimized for SSDs (https://github.com/dgraph-io/badger)
    #   memory: in-memory version of badger
  store: {{ .Datastore.Store }}
  badger:
    path: {{ .Datastore.Badger.Path }}
  # memory:
  #    size: {{ .Datastore.Memory.Size }}

api:
  # Listening address of the (HTTP API) GraphQL query endpoint
  address: {{ .API.Address }}

net:
  p2pdisabled: {{ .Net.P2PDisabled }}
  p2paddress: {{ .Net.P2PAddress }}
  rpcaddress: {{ .Net.RPCAddress }}
  # gRPC server address
  tcpaddress: {{ .Net.TCPAddress }}
  # Time duration after which a RPC connection to a peer times out
  rpctimeout: {{ .Net.RPCTimeout }}
  # Whether the node has pubsub enabled or not
  pubsub: {{ .Net.PubSubEnabled }}
  # Enable libp2p's Circuit relay transport protocol https://docs.libp2p.io/concepts/circuit-relay/
  relay: {{ .Net.RelayEnabled }}
  # List of peers to boostrap with, specified as multiaddresses (https://docs.libp2p.io/concepts/addressing/)
  peers: {{ .Net.Peers }}

logging:
  # Log level. Options are debug, info, warn, error, fatal
  level: {{ .Logging.Level }}
  # Include stacktrace in error and fatal logs
  stacktrace: {{ .Logging.Stacktrace }}
  # Supported log formats are json, csv
  format: {{ .Logging.Format }}
  # Where the log output is written to
  outputpath: {{ .Logging.OutputPath }}
`
