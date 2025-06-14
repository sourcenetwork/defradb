// Copyright 2022 Democratized Data Foundation
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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/telemetry"
	"github.com/sourcenetwork/defradb/keyring"
	netConfig "github.com/sourcenetwork/defradb/net/config"
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
 - generates temporary node identity if keyring is disabled`

func MakeStartCommand() *cobra.Command {
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
			if err := createConfig(rootdir, cmd.Flags()); err != nil {
				return err
			}
			return setContextConfig(cmd)
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

			opts := []node.Option{
				node.WithDisableP2P(cfg.GetBool("net.p2pDisabled")),
				node.WithSourceHubChainID(cfg.GetString("acp.dac.sourceHub.ChainID")),
				node.WithSourceHubGRPCAddress(cfg.GetString("acp.dac.sourceHub.GRPCAddress")),
				node.WithSourceHubCometRPCAddress(cfg.GetString("acp.dac.sourceHub.CometRPCAddress")),
				node.WithLensRuntime(node.LensRuntimeType(cfg.GetString("lens.runtime"))),
				node.WithEnableDevelopment(cfg.GetBool("development")),
				// store options
				node.WithStorePath(cfg.GetString("datastore.badger.path")),
				node.WithBadgerInMemory(cfg.GetString("datastore.store") == configStoreMemory),
				// db options
				db.WithMaxRetries(cfg.GetInt("datastore.MaxTxnRetries")),
				db.WithRetryInterval(replicatorRetryIntervals),
				// net node options
				netConfig.WithListenAddresses(cfg.GetStringSlice("net.p2pAddresses")...),
				netConfig.WithEnablePubSub(cfg.GetBool("net.pubSubEnabled")),
				netConfig.WithEnableRelay(cfg.GetBool("net.relayEnabled")),
				netConfig.WithBootstrapPeers(cfg.GetStringSlice("net.peers")...),
				// http server options
				http.WithAddress(cfg.GetString("api.address")),
				http.WithAllowedOrigins(cfg.GetStringSlice("api.allowed-origins")...),
				http.WithTLSCertPath(cfg.GetString("api.pubKeyPath")),
				http.WithTLSKeyPath(cfg.GetString("api.privKeyPath")),
			}

			if cfg.GetString("datastore.store") != configStoreMemory {
				rootDir := mustGetContextRootDir(cmd)
				// TODO-ACP: Infuture when we add support for the --no-acp flag when admin signatures are in,
				// we can allow starting of db without acp. Currently that can only be done programmatically.
				// https://github.com/sourcenetwork/defradb/issues/2271
				opts = append(opts, node.WithDocumentACPPath(rootDir))
			}

			documentACPType := cfg.GetString("acp.dac.type")
			if documentACPType != "" {
				opts = append(opts, node.WithDocumentACPType(node.DocumentACPType(documentACPType)))
			}

			if !cfg.GetBool("keyring.disabled") {
				kr, err := openKeyring(cmd)
				if err != nil {
					return err
				}
				opts, err = getOrCreatePeerKey(kr, opts)
				if err != nil {
					return err
				}

				if !cfg.GetBool("datastore.noencryption") {
					opts, err = getOrCreateEncryptionKey(kr, opts)
					if err != nil {
						return err
					}
				}

				opts, err = getOrCreateIdentity(kr, opts, cfg)
				if err != nil {
					return err
				}

				// setup the sourcehub transaction signer
				sourceHubKeyName := cfg.GetString("acp.dac.sourceHub.KeyName")
				if sourceHubKeyName != "" {
					signer, err := keyring.NewTxSignerFromKeyringKey(kr, sourceHubKeyName)
					if err != nil {
						return err
					}
					opts = append(opts, node.WithTxnSigner(immutable.Some[node.TxSigner](signer)))
				}
			}

			opts = append(opts, db.WithEnabledSigning(!cfg.GetBool("datastore.nosigning")))

			isDevMode := cfg.GetBool("development")
			http.IsDevMode = isDevMode
			if isDevMode {
				cmd.Printf(devModeBanner)
				if cfg.GetBool("keyring.disabled") {
					var err error
					// Generate an ephemeral identity for the node
					// TODO: we want to persist this identity so we can restart the node with the same identity
					// even in development mode. https://github.com/sourcenetwork/defradb/issues/3148
					ident, err := generateIdentity(cfg.GetString("datastore.defaultkeytype"))
					if err != nil {
						return err
					}
					opts = append(opts, db.WithNodeIdentity(ident))
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

			n, err := node.New(cmd.Context(), opts...)
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
				if errors.Is(err, node.ErrPurgeWithDevModeDisabled) {
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
	cfg := defaultConfig()
	cmd.PersistentFlags().StringArray(
		"peers",
		cfg.GetStringSlice(configFlags["peers"]),
		"List of peers to connect to",
	)
	cmd.PersistentFlags().Int(
		"max-txn-retries",
		cfg.GetInt(configFlags["max-txn-retries"]),
		"Specify the maximum number of retries per transaction",
	)
	cmd.PersistentFlags().String(
		"store",
		cfg.GetString(configFlags["store"]),
		"Specify the datastore to use (supported: badger, memory)",
	)
	cmd.PersistentFlags().Int(
		"valuelogfilesize",
		cfg.GetInt(configFlags["valuelogfilesize"]),
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)
	cmd.PersistentFlags().StringSlice(
		"p2paddr",
		cfg.GetStringSlice(configFlags["p2paddr"]),
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)
	cmd.PersistentFlags().Bool(
		"no-p2p",
		cfg.GetBool(configFlags["no-p2p"]),
		"Disable the peer-to-peer network synchronization system",
	)
	cmd.PersistentFlags().StringArray(
		"allowed-origins",
		cfg.GetStringSlice(configFlags["allowed-origins"]),
		"List of origins to allow for CORS requests",
	)
	cmd.PersistentFlags().String(
		"pubkeypath",
		cfg.GetString(configFlags["pubkeypath"]),
		"Path to the public key for tls",
	)
	cmd.PersistentFlags().String(
		"privkeypath",
		cfg.GetString(configFlags["privkeypath"]),
		"Path to the private key for tls",
	)
	cmd.PersistentFlags().Bool(
		"development",
		cfg.GetBool(configFlags["development"]),
		developmentDescription,
	)
	cmd.Flags().Bool(
		"no-encryption",
		cfg.GetBool(configFlags["no-encryption"]),
		"Skip generating an encryption key. Encryption at rest will be disabled. WARNING: This cannot be undone.")
	cmd.PersistentFlags().Bool(
		"no-telemetry",
		cfg.GetBool(configFlags["no-telemetry"]),
		"Disables telemetry reporting. Telemetry is only enabled in builds that use the telemetry flag.",
	)
	cmd.Flags().Bool(
		"no-signing",
		cfg.GetBool(configFlags["no-signing"]),
		"Disable signing of commits.")
	cmd.Flags().String(
		"default-key-type",
		cfg.GetString(configFlags["default-key-type"]),
		"Default key type to generate new node identity if one doesn't exist in the keyring. "+
			"Valid values are 'secp256k1' and 'ed25519'. "+
			"If not specified, the default key type will be 'secp256k1'.")
	cmd.PersistentFlags().String(
		"dac-type",
		cfg.GetString(configFlags["acp.dac.type"]),
		"Specify the document acp engine to use (supported: none (default), local, source-hub)")
	cmd.PersistentFlags().IntSlice(
		"replicator-retry-intervals",
		cfg.GetIntSlice(configFlags["replicator-retry-intervals"]),
		"Retry intervals for the replicator. Format is a comma-separated list of whole number seconds. "+
			"Example: 10,20,40,80,160,320",
	)
	return cmd
}

func getOrCreateEncryptionKey(kr keyring.Keyring, opts []node.Option) ([]node.Option, error) {
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
			return nil, err
		}
		log.Info("generated encryption key")
	}
	opts = append(opts, node.WithBadgerEncryptionKey(encryptionKey))
	return opts, nil
}

func getOrCreatePeerKey(kr keyring.Keyring, opts []node.Option) ([]node.Option, error) {
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
		log.Info("generated peer key")
	} else if err != nil {
		return nil, err
	}
	return append(opts, netConfig.WithPrivateKey(peerKey)), nil
}

func getOrCreateIdentity(kr keyring.Keyring, opts []node.Option, cfg *viper.Viper) ([]node.Option, error) {
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
		rawKey := ident.PrivateKey.Raw()
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
		return getOrCreateIdentity(kr, opts, cfg)
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

	return append(opts, db.WithNodeIdentity(ident)), nil
}

func generateIdentity(keyType string) (identity.Identity, error) {
	privateKey, err := crypto.GenerateKey(crypto.KeyType(keyType))
	if err != nil {
		return identity.Identity{}, err
	}

	nodeIdentity, err := identity.FromPrivateKey(privateKey)
	if err != nil {
		return identity.Identity{}, err
	}

	return nodeIdentity, nil
}
