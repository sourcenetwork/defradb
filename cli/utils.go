// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/keyring"
)

const (
	peerKeyName         = "peer-key"
	encryptionKeyName   = "encryption-key"
	nodeIdentityKeyName = "node-identity-key"
)

type contextKey string

var (
	// cfgContextKey is the context key for the config.
	cfgContextKey = contextKey("cfg")
	// rootDirContextKey is the context key for the root directory.
	rootDirContextKey = contextKey("rootDir")
	// dbContextKey is the context key for the client.DB
	dbContextKey = contextKey("db")
	// colContextKey is the context key for the client.Collection
	//
	// If a transaction exists, all operations will be executed
	// in the current transaction context.
	colContextKey = contextKey("col")
)

const (
	// authTokenExpiration is the default expiration time for auth tokens.
	authTokenExpiration = time.Minute * 15
)

// mustGetContextDB returns the db for the current command context.
//
// If a db is not set in the current context this function panics.
func mustGetContextDB(cmd *cobra.Command) client.DB {
	return cmd.Context().Value(dbContextKey).(client.DB) //nolint:forcetypeassert
}

// mustGetContextStore returns the store for the current command context.
//
// If a store is not set in the current context this function panics.
func mustGetContextStore(cmd *cobra.Command) client.Store {
	return cmd.Context().Value(dbContextKey).(client.Store) //nolint:forcetypeassert
}

// mustGetContextP2P returns the p2p implementation for the current command context.
//
// If a p2p implementation is not set in the current context this function panics.
func mustGetContextP2P(cmd *cobra.Command) client.P2P {
	return cmd.Context().Value(dbContextKey).(client.P2P) //nolint:forcetypeassert
}

// mustGetContextHTTP returns the http client for the current command context.
//
// If http client is not set in the current context this function panics.
func mustGetContextHTTP(cmd *cobra.Command) *http.Client {
	return cmd.Context().Value(dbContextKey).(*http.Client) //nolint:forcetypeassert
}

// mustGetContextConfig returns the config for the current command context.
//
// If a config is not set in the current context this function panics.
func mustGetContextConfig(cmd *cobra.Command) *viper.Viper {
	return cmd.Context().Value(cfgContextKey).(*viper.Viper) //nolint:forcetypeassert
}

// mustGetContextRootDir returns the rootdir for the current command context.
//
// If a rootdir is not set in the current context this function panics.
func mustGetContextRootDir(cmd *cobra.Command) string {
	return cmd.Context().Value(rootDirContextKey).(string) //nolint:forcetypeassert
}

// tryGetContextCollection returns the collection for the current command context
// and a boolean indicating if the collection was set.
func tryGetContextCollection(cmd *cobra.Command) (client.Collection, bool) {
	col, ok := cmd.Context().Value(colContextKey).(client.Collection)
	return col, ok
}

// setContextDB sets the db for the current command context.
func setContextDB(cmd *cobra.Command) error {
	cfg := mustGetContextConfig(cmd)
	db, err := http.NewClient(cfg.GetString("api.address"))
	if err != nil {
		return err
	}
	ctx := context.WithValue(cmd.Context(), dbContextKey, db)
	cmd.SetContext(ctx)
	return nil
}

// setContextConfig sets teh config for the current command context.
func setContextConfig(cmd *cobra.Command) error {
	rootdir := mustGetContextRootDir(cmd)
	cfg, err := loadConfig(rootdir, cmd.Flags())
	if err != nil {
		return err
	}
	ctx := context.WithValue(cmd.Context(), cfgContextKey, cfg)
	cmd.SetContext(ctx)
	return nil
}

// setContextTransaction sets the transaction for the current command context.
func setContextTransaction(cmd *cobra.Command, txId uint64) error {
	if txId == 0 {
		return nil
	}
	cfg := mustGetContextConfig(cmd)
	tx, err := http.NewTransaction(cfg.GetString("api.address"), txId)
	if err != nil {
		return err
	}
	ctx := db.SetContextTxn(cmd.Context(), tx)
	cmd.SetContext(ctx)
	return nil
}

// setContextIdentity sets the identity for the current command context.
func setContextIdentity(cmd *cobra.Command, privateKeyHex string) error {
	if privateKeyHex == "" {
		return nil
	}
	data, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return err
	}

	cfg := mustGetContextConfig(cmd)

	sourcehubAddressString := cfg.GetString("acp.sourceHub.address")
	var sourcehubAddress immutable.Option[string]
	if sourcehubAddressString != "" {
		sourcehubAddress = immutable.Some(sourcehubAddressString)
	}

	privKey := secp256k1.PrivKeyFromBytes(data)
	ident, err := acpIdentity.FromPrivateKey(privKey)
	if err != nil {
		return err
	}
	err = ident.UpdateToken(
		authTokenExpiration,
		immutable.Some(cfg.GetString("api.address")),
		sourcehubAddress)
	if err != nil {
		return err
	}

	ctx := identity.WithContext(cmd.Context(), immutable.Some(ident))
	cmd.SetContext(ctx)
	return nil
}

// setContextRootDir sets the rootdir for the current command context.
func setContextRootDir(cmd *cobra.Command) error {
	rootdir, err := cmd.Root().PersistentFlags().GetString("rootdir")
	if err != nil {
		return err
	}
	if rootdir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		rootdir = filepath.Join(home, ".defradb")
	}
	ctx := context.WithValue(cmd.Context(), rootDirContextKey, rootdir)
	cmd.SetContext(ctx)
	return nil
}

// openKeyring opens the keyring for the current environment.
func openKeyring(cmd *cobra.Command) (keyring.Keyring, error) {
	cfg := mustGetContextConfig(cmd)
	backend := cfg.Get("keyring.backend")
	if backend == "system" {
		return keyring.OpenSystemKeyring(cfg.GetString("keyring.namespace")), nil
	}
	if backend != "file" {
		log.Info("keyring defaulted to file backend")
	}
	path := cfg.GetString("keyring.path")
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	secret := []byte(cfg.GetString("keyring.secret"))
	if len(secret) == 0 {
		return nil, ErrMissingKeyringSecret
	}
	return keyring.OpenFileKeyring(path, secret)
}

func writeJSON(cmd *cobra.Command, out any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
