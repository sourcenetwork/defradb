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
	"bytes"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli/config"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/telemetry"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/version"
)

const devModeBanner = `
******************************************
**     DEVELOPMENT MODE IS ENABLED      **
** ------------------------------------ **
**   if this is a production database   **
** disable development mode and restart **
**   or you may risk losing all data    **
******************************************

`

const developmentDescription = `Enables a set of features that make development easier but should not be enabled ` +
	`in production:
 - allows purging of all persisted data
 - generates temporary node identity if one doesn't exist in the keyring`

func MakeStartCommand(ctx context.Context) *cobra.Command {
	var identity string
	var enableNAC bool
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a DefraDB node",
		Long:  "Start a DefraDB node.",
		// Load the root config if it exists, otherwise create it.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			rootdir := mustGetContextRootDir(cmd)
			if err := config.CreateConfig(rootdir, cmd.Flags()); err != nil {
				return err
			}
			if err := setContextConfig(cmd); err != nil {
				return err
			}
			if err := setContextIdentity(cmd, identity); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mustGetContextConfig(cmd)

			// Parse the retry intervals from the config files from a slice of ints to a slice of time.Durations
			replicatorRetryIntervals := []time.Duration{}
			for _, interval := range cfg.GetIntSlice("replicator.retryintervals") {
				if interval <= 0 {
					return ErrNegativeReplicatorRetryIntervals
				}
				replicatorRetryIntervals = append(replicatorRetryIntervals, time.Duration(interval)*time.Second)
			}

			inMem := cfg.GetString("datastore.store") == config.ConfigStoreMemory

			opts := options.Node().
				SetEnableDevelopment(cfg.GetBool("development")).
				SetDisableP2P(cfg.GetBool("net.p2pDisabled"))
			opts.Store().
				SetPath(cfg.GetString("datastore.badger.path")).
				SetBadgerInMemory(inMem)
			opts.DB().
				SetMaxTxnRetries(cfg.GetInt("datastore.MaxTxnRetries")).
				SetRetryIntervals(replicatorRetryIntervals).
				SetLensRuntime(options.NodeLensRuntimeType(cfg.GetString("lens.runtime")))
			opts.P2P().
				SetListenAddresses(cfg.GetStringSlice("net.p2pAddresses")...).
				SetEnablePubSub(cfg.GetBool("net.pubSubEnabled")).
				SetEnableRelay(cfg.GetBool("net.relayEnabled")).
				SetBootstrapPeers(cfg.GetStringSlice("net.peers")...)
			opts.HTTP().
				SetAddress(cfg.GetString("api.address")).
				SetAllowedOrigins(cfg.GetStringSlice("api.allowed-origins")...).
				SetCertPath(cfg.GetString("api.pubKeyPath")).
				SetKeyPath(cfg.GetString("api.privKeyPath"))
			opts.DocumentACP().
				SetChainID(cfg.GetString("acp.document.sourceHub.ChainID")).
				SetGRPCAddress(cfg.GetString("acp.document.sourceHub.GRPCAddress")).
				SetCometRPCAddress(cfg.GetString("acp.document.sourceHub.CometRPCAddress"))
			opts.NodeACP().
				SetEnabled(enableNAC)

			if !inMem {
				rootDir := mustGetContextRootDir(cmd)
				opts.DocumentACP().SetPath(rootDir).
					NodeACP().SetPath(rootDir)
			}

			if enableNAC && identity == "" {
				return client.ErrCanNotStartNACWithoutIdentity
			}

			documentACPType := cfg.GetString("acp.document.type")
			if documentACPType != "" {
				opts.DocumentACP().SetType(options.NodeDocumentACPType(documentACPType))
			}

			if !cfg.GetBool("keyring.disabled") {
				kr, err := openKeyring(cmd)
				if err != nil {
					return err
				}
				peerKey, err := getOrCreatePeerKey(kr)
				if err != nil {
					return err
				}
				opts.P2P().SetPrivateKey(peerKey)

				if !cfg.GetBool("datastore.noencryption") {
					encKey, err := getOrCreateEncryptionKey(kr)
					if err != nil {
						return err
					}
					opts.Store().SetBadgerEncryptionKey(encKey)
				}

				if !cfg.GetBool("datastore.nosearchableencryption") {
					seKey, err := getOrCreateSearchableEncryptionKey(kr)
					if err != nil {
						return err
					}
					opts.DB().SetSearchableEncryptionKey(seKey)
				}

				ident, err := getOrCreateIdentity(kr, cfg)
				if err != nil {
					return err
				}
				opts.DB().SetNodeIdentity(ident)

				// setup the sourcehub transaction signer
				sourceHubKeyName := cfg.GetString("acp.document.sourceHub.KeyName")
				if sourceHubKeyName != "" {
					signer, err := keyring.NewTxSignerFromKeyringKey(kr, sourceHubKeyName)
					if err != nil {
						return err
					}
					opts.DocumentACP().SetTxnSigner(signer)
				}
			}

			opts.DB().SetEnableSigning(!cfg.GetBool("datastore.nosigning"))

			isDevMode := cfg.GetBool("development")
			if isDevMode {
				cmd.Printf(devModeBanner)
				if cfg.GetBool("keyring.disabled") {
					// Generate an ephemeral identity for the node
					ident, err := generateIdentity(cfg.GetString("datastore.defaultkeytype"))
					if err != nil {
						return err
					}
					opts.DB().SetNodeIdentity(ident)
				}
			}

			if !cfg.GetBool("no-telemetry") {
				ver, err := version.NewDefraVersion()
				if err != nil {
					return err
				}
				err = telemetry.ConfigureTelemetry(cmd.Context(), ver.String())
				if err != nil {
					log.ErrorContextE(cmd.Context(), "failed to configure telemetry", err)
				}
			}

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

			n, err := node.New(cmd.Context(), opts)
			if err != nil {
				return err
			}
			log.InfoContext(cmd.Context(), "Starting DefraDB")
			if err := n.Start(cmd.Context()); err != nil {
				return err
			}
			// If the context has a messageChans defined, we pass along the relevant information.
			// For now this is mostly useful for the CLI integration tests.
			messageChans, ok := node.TryGetContextMessageChans(cmd.Context())
			if ok && messageChans.APIURL != nil {
				messageChans.APIURL <- n.APIURL
				close(messageChans.APIURL)
			}

		RESTART:
			// after a restart we need to resubscribe
			purgeSub, err := n.DB.Events().Subscribe(event.PurgeName)
			if err != nil {
				return err
			}

		SELECT:
			select {
			case <-purgeSub.Message():
				log.InfoContext(cmd.Context(), "Received purge event; restarting...")

				err := n.PurgeAndRestart(cmd.Context())
				if err != nil {
					log.ErrorContextE(cmd.Context(), "failed to purge", err)
				}
				if err == nil {
					goto RESTART
				}
				if errors.Is(err, client.ErrOperationRequiresDeveloperMode) {
					goto SELECT
				}

			case <-cmd.Context().Done():
				log.InfoContext(cmd.Context(), "Received context cancellation; shutting down...")

			case <-signalCh:
				log.InfoContext(cmd.Context(), "Received interrupt; shutting down...")
			}

			return n.Close(cmd.Context())
		},
	}
	// set default flag values from config
	cfg := config.DefaultConfig()
	cmd.PersistentFlags().StringArray(
		"peers",
		cfg.GetStringSlice(config.ConfigFlags["peers"]),
		"List of peers to connect to",
	)
	cmd.PersistentFlags().Int(
		"max-txn-retries",
		cfg.GetInt(config.ConfigFlags["max-txn-retries"]),
		"Specify the maximum number of retries per transaction",
	)
	cmd.PersistentFlags().String(
		"store",
		cfg.GetString(config.ConfigFlags["store"]),
		"Specify the datastore to use (supported: badger, memory)",
	)
	cmd.PersistentFlags().Int(
		"valuelogfilesize",
		cfg.GetInt(config.ConfigFlags["valuelogfilesize"]),
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	cmd.PersistentFlags().StringSlice(
		"p2paddr",
		cfg.GetStringSlice(config.ConfigFlags["p2paddr"]),
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)
	cmd.PersistentFlags().Bool(
		"no-p2p",
		cfg.GetBool(config.ConfigFlags["no-p2p"]),
		"Disable the peer-to-peer network synchronization system",
	)
	cmd.PersistentFlags().StringArray(
		"allowed-origins",
		cfg.GetStringSlice(config.ConfigFlags["allowed-origins"]),
		"List of origins to allow for CORS requests",
	)
	cmd.PersistentFlags().String(
		"pubkeypath",
		cfg.GetString(config.ConfigFlags["pubkeypath"]),
		"Path to the public key for tls",
	)
	cmd.PersistentFlags().String(
		"privkeypath",
		cfg.GetString(config.ConfigFlags["privkeypath"]),
		"Path to the private key for tls",
	)
	cmd.PersistentFlags().Bool(
		"development",
		cfg.GetBool(config.ConfigFlags["development"]),
		developmentDescription,
	)
	cmd.Flags().Bool(
		"no-encryption",
		cfg.GetBool(config.ConfigFlags["no-encryption"]),
		"Skip generating an encryption key. Encryption at rest will be disabled. WARNING: This cannot be undone.")
	cmd.PersistentFlags().Bool(
		"no-telemetry",
		cfg.GetBool(config.ConfigFlags["no-telemetry"]),
		"Disables telemetry reporting. Telemetry is only enabled in builds that use the telemetry flag.",
	)
	cmd.Flags().Bool(
		"no-signing",
		cfg.GetBool(config.ConfigFlags["no-signing"]),
		"Disable signing of commits.")
	cmd.Flags().String(
		"default-key-type",
		cfg.GetString(config.ConfigFlags["default-key-type"]),
		"Default key type to generate new node identity if one doesn't exist in the keyring. "+
			"Valid values are 'secp256k1' and 'ed25519'. "+
			"If not specified, the default key type will be 'secp256k1'.")
	cmd.Flags().Bool(
		"no-searchable-encryption",
		cfg.GetBool(config.ConfigFlags["no-searchable-encryption"]),
		"Skip generating a searchable encryption key. Searchable encryption will be disabled.")
	cmd.PersistentFlags().StringVarP(
		&identity,
		"identity",
		"i",
		"",
		"Hex formatted private key used to authenticate with ACP",
	)
	cmd.PersistentFlags().BoolVar(
		&enableNAC,
		"node-acp-enable",
		false,
		"Enable the node access control system.",
	)
	cmd.PersistentFlags().String(
		"document-acp-type",
		cfg.GetString(config.ConfigFlags["document-acp-type"]),
		"Specify the document acp engine to use (supported: none (default), local, source-hub)")
	cmd.PersistentFlags().IntSlice(
		"replicator-retry-intervals",
		cfg.GetIntSlice(config.ConfigFlags["replicator-retry-intervals"]),
		"Retry intervals for the replicator. Format is a comma-separated list of whole number seconds. "+
			"Example: 10,20,40,80,160,320",
	)
	return cmd
}

func getOrCreateEncryptionKey(kr keyring.Keyring) ([]byte, error) {
	encryptionKey, err := kr.Get(encryptionKeyName)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			return nil, err
		}
		encryptionKey, err = crypto.GenerateAES256()
		if err != nil {
			return nil, err
		}
		err = kr.Set(encryptionKeyName, encryptionKey)
		if err != nil {
			return nil, NewErrStoreEncryptionKey(err)
		}
	}
	return encryptionKey, nil
}

// getOrCreateSearchableEncryptionKey generates or retrieves the searchable encryption key
// from the keyring.
func getOrCreateSearchableEncryptionKey(kr keyring.Keyring) ([]byte, error) {
	seKey, err := kr.Get(searchableEncryptionKeyName)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			return nil, err
		}
		seKey, err = crypto.GenerateAES256()
		if err != nil {
			return nil, err
		}
		err = kr.Set(searchableEncryptionKeyName, seKey)
		if err != nil {
			return nil, err
		}
	}
	return seKey, nil
}

func getOrCreatePeerKey(kr keyring.Keyring) ([]byte, error) {
	peerKey, err := kr.Get(peerKeyName)
	if err != nil && errors.Is(err, keyring.ErrNotFound) {
		peerKey, err = crypto.GenerateEd25519()
		if err != nil {
			return nil, err
		}
		err = kr.Set(peerKeyName, peerKey)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return peerKey, nil
}

func getOrCreateIdentity(kr keyring.Keyring, cfg *viper.Viper) (identity.Identity, error) {
	identityBytes, err := kr.Get(nodeIdentityKeyName)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			return nil, err
		}
		keyType := cfg.GetString("datastore.defaultkeytype")
		ident, err := generateIdentity(keyType)
		if err != nil {
			return nil, err
		}
		rawKey := ident.PrivateKey().Raw()
		// Make sure the outerscope knows about the newly created identity
		identityBytes = append([]byte(keyType+":"), rawKey...)
		err = kr.Set(nodeIdentityKeyName, identityBytes)
		if err != nil {
			return nil, err
		}
	}

	sepPos := bytes.Index(identityBytes, []byte(":"))
	// the separator might not exist, because of the old format of storing it
	// we turn it into the new format and try again
	if sepPos == -1 {
		identityBytes = append([]byte(crypto.KeyTypeSecp256k1+":"), identityBytes...)
		err = kr.Set(nodeIdentityKeyName, identityBytes)
		if err != nil {
			return nil, err
		}
		return getOrCreateIdentity(kr, cfg)
	}
	keyType := string(identityBytes[:sepPos])
	privateKey, err := crypto.PrivateKeyFromBytes(crypto.KeyType(keyType), identityBytes[sepPos+1:])
	if err != nil {
		return nil, err
	}
	ident, err := identity.FromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func generateIdentity(keyType string) (identity.FullIdentity, error) {
	privateKey, err := crypto.GenerateKey(crypto.KeyType(keyType))
	if err != nil {
		return nil, err
	}

	nodeIdentity, err := identity.FromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return nodeIdentity, nil
}
