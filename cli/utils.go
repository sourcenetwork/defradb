// Copyright 2025 Democratized Data Foundation
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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli/config"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/datastore"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
)

const (
	peerKeyName                 = "peer-key"
	encryptionKeyName           = "encryption-key"
	nodeIdentityKeyName         = "node-identity-key"
	searchableEncryptionKeyName = "searchable-encryption-key"
)

type contextKey string

var (
	// cfgContextKey is the context key for the config.
	cfgContextKey = contextKey("cfg")
	// rootDirContextKey is the context key for the root directory.
	rootDirContextKey = contextKey("rootDir")
	// clientContextKey is the context key for the cliclient.TxnStore
	clientContextKey = contextKey("client")
	// colContextKey is the context key for the cliClient.Collection
	//
	// If a transaction exists, all operations will be executed
	// in the current transaction context.
	colContextKey = contextKey("col")
)

const (
	// authTokenExpiration is the default expiration time for auth tokens.
	authTokenExpiration = time.Minute * 15
)

// mustGetContextCLIClient returns the CLI for the current command context.
//
// If a CLI is not set in the current context this function panics.
func mustGetContextCLIClient(cmd *cobra.Command) CLI {
	return cmd.Context().Value(clientContextKey).(CLI) //nolint:forcetypeassert
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

// setContextClient sets the db for the current command context.
func setContextClient(cmd *cobra.Command) error {
	cfg := mustGetContextConfig(cmd)
	client, err := http.NewClient(cfg.GetString("api.address"))
	if err != nil {
		return err
	}
	ctx := context.WithValue(cmd.Context(), clientContextKey, client)
	cmd.SetContext(ctx)
	return nil
}

// setContextConfig sets the config for the current command context.
func setContextConfig(cmd *cobra.Command) error {
	rootdir := mustGetContextRootDir(cmd)
	cfg, err := config.LoadConfig(rootdir, cmd.Flags())
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
	ctx := datastore.CtxSetFromClientTxn(cmd.Context(), tx)
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

	sourcehubAddressString := cfg.GetString("acp.document.sourceHub.address")
	var sourcehubAddress immutable.Option[string]
	if sourcehubAddressString != "" {
		sourcehubAddress = immutable.Some(sourcehubAddressString)
	}

	privKey := secp256k1.PrivKeyFromBytes(data)
	ident, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
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

	ctx := iIdentity.WithContext(cmd.Context(), immutable.Some[acpIdentity.Identity](ident))
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
		rootdir = node.GetDefaultStorePath()
	}
	ctx := context.WithValue(cmd.Context(), rootDirContextKey, rootdir)
	cmd.SetContext(ctx)
	return nil
}

// openKeyring opens the keyring for the current environment.
func openKeyring(cmd *cobra.Command) (keyring.Keyring, error) {
	cfg := mustGetContextConfig(cmd)
	backend := cfg.Get("keyring.backend")
	if backend == keyring.KeyringBackendSystem {
		return keyring.OpenSystemKeyring(cfg.GetString("keyring.namespace")), nil
	}
	if backend != keyring.KeyringBackendFile {
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

// withExampleRegistry injects an ExampleRegitry into the context.
// This is primarily only needed by the tests but needs to exist
// in the main CLI package
func withExampleRegistry(ctx context.Context, registry *exampleRegistry) context.Context {
	return context.WithValue(ctx, exampleRegistryCtxKey{}, registry)
}

type exampleRegistry struct {
	examples map[string]string
}

func newExampleRegistry() *exampleRegistry {
	return &exampleRegistry{
		examples: make(map[string]string),
	}
}

// EmbedCLIExample will embed the given CLI usage example into the provided cobra command `Example` field.
// Most notably, it will also register the example with a `exampleRegistry` if it exists in the provided
// context. This enables the `exampleRegistry` to expose the examples in a programatic way that can be
// accessed by the test suite.
//
// This means we can maintain correctness between our CLI examples, and docs, while also programatically
// validating the examples, so they can't drift from the implementation. Note, this is *only* validating
// the commands, flags, and arguments. Its not actually running the full command execution.
//
// It *must* be called *after* the `cmd` object has been defined. Beyond that, it doesn't matter if
// its before or after the flag definitions. It is reccomended to immedietly follow the command definition
// for clarity/consistency.
//
// You may use any "name", such as a short title or even a somewhat longer description. It is used for
// uniqueness and for error reporting if an example fails. The actual name used on the registry is
// combined with the cmd.Short (if it exists).
//
// Check out `cli/cli_test.go:TestCLIExamples()` for the test consuming side of the example registry.
func EmbedCLIExample(ctx context.Context, cmd *cobra.Command, name, usage string) {
	exampleString := cliExampleToString(name, usage)
	if cmd.Example != "" {
		cmd.Example += "\n\n"
	}
	cmd.Example += exampleString

	cmdName := cmd.Short
	if cmdName == "" {
		cmdName = cmd.Name()
	}
	exampleName := cmdName + "/" + name
	registerCLIExample(ctx, exampleName, usage)
}

type exampleRegistryCtxKey struct{}

func cliExampleToString(name, usage string) string {
	// this is intentionally formatted this way, including
	// the 2 white spaces at the start/end of the lines
	return fmt.Sprintf(`%s:  
  %s`, name, usage)
}

func registerCLIExample(ctx context.Context, name, usage string) {
	registry, ok := ctx.Value(exampleRegistryCtxKey{}).(*exampleRegistry)
	if !ok {
		return
	}

	_, exists := registry.examples[name]
	if exists {
		panic("CLI example with the same name already exists: " + name)
	}

	if strings.Contains(usage, " | ") {
		usageParts := strings.Split(usage, "|")
		usage = usageParts[1]
	}
	registry.examples[name] = strings.ReplaceAll(usage, "\\\n", "")
}

func validateCLIArgs(cmd *cobra.Command, args []string) error {
	cmd, args, err := cmd.Find(args)
	if err != nil {
		return err
	}

	if !cmd.Runnable() {
		return fmt.Errorf("command isn't runnable: %s", cmd.Name())
	}

	flags := cmd.Flags()
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	remainingArgs := flags.Args()

	if cmd.Args != nil {
		if err := cmd.Args(cmd, remainingArgs); err != nil {
			return err
		}
	}

	if err := cmd.ValidateRequiredFlags(); err != nil {
		return err
	}

	return nil
}
