// Copyright 2024 Democratized Data Foundation
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
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/cli/config"
)

// writeDefaultCerts writes placeholder certificate and/or key files into the
// default <rootdir>/certs directory and returns their paths. Passing false for
// either skips writing that file (to exercise the partial-pair case).
func writeDefaultCerts(t *testing.T, rootdir string, writeCert, writeKey bool) (string, string) {
	certDir := filepath.Join(rootdir, "certs")
	require.NoError(t, os.MkdirAll(certDir, 0755))
	certPath := filepath.Join(certDir, "server.crt")
	keyPath := filepath.Join(certDir, "server.key")
	if writeCert {
		require.NoError(t, os.WriteFile(certPath, []byte("test-cert"), 0644))
	}
	if writeKey {
		require.NoError(t, os.WriteFile(keyPath, []byte("test-key"), 0600))
	}
	return certPath, keyPath
}

func TestCreateConfig(t *testing.T) {
	rootdir := t.TempDir()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	err := config.CreateConfig(rootdir, flags)
	require.NoError(t, err)

	// ensure no errors when config already exists
	err = config.CreateConfig(rootdir, flags)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(rootdir, "config.yaml"))
}

func TestLoadConfigNotExist(t *testing.T) {
	rootdir := t.TempDir()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	cfg, err := config.LoadConfig(rootdir, flags)
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.GetInt("datastore.maxtxnretries"))
	assert.Equal(t, filepath.Join(rootdir, "data"), cfg.GetString("datastore.path"))
	assert.Equal(t, 1<<30, cfg.GetInt("datastore.badger.valuelogfilesize"))
	assert.Equal(t, "badger", cfg.GetString("datastore.store"))

	assert.Equal(t, "127.0.0.1:9181", cfg.GetString("api.address"))
	assert.Equal(t, "", cfg.GetString("api.audience"))
	assert.Equal(t, []string{}, cfg.GetStringSlice("api.allowed-origins"))
	assert.Equal(t, "", cfg.GetString("api.pubkeypath"))
	assert.Equal(t, "", cfg.GetString("api.privkeypath"))

	assert.Equal(t, false, cfg.GetBool("net.p2pdisabled"))
	assert.Equal(t, []string{"/ip4/127.0.0.1/tcp/9171"}, cfg.GetStringSlice("net.p2paddresses"))
	assert.Equal(t, true, cfg.GetBool("net.pubsubenabled"))
	assert.Equal(t, false, cfg.GetBool("net.relay"))
	assert.Equal(t, []string{}, cfg.GetStringSlice("net.peers"))

	assert.Equal(t, "info", cfg.GetString("log.level"))
	assert.Equal(t, "stderr", cfg.GetString("log.output"))
	assert.Equal(t, "text", cfg.GetString("log.format"))
	assert.Equal(t, false, cfg.GetBool("log.stacktrace"))
	assert.Equal(t, false, cfg.GetBool("log.source"))
	assert.Equal(t, "", cfg.GetString("log.overrides"))
	assert.Equal(t, false, cfg.GetBool("log.colordisabled"))

	assert.Equal(t, filepath.Join(rootdir, "keys"), cfg.GetString("keyring.path"))
	assert.Equal(t, false, cfg.GetBool("keyring.disabled"))
	assert.Equal(t, "defradb", cfg.GetString("keyring.namespace"))
	assert.Equal(t, "file", cfg.GetString("keyring.backend"))

	assert.Equal(t, false, cfg.GetBool("development"))
	assert.Equal(t, false, cfg.GetBool("telemetry.disabled"))
}

func TestLoadConfigAutoEnablesTLSWhenDefaultCertsPresent(t *testing.T) {
	rootdir := t.TempDir()
	certPath, keyPath := writeDefaultCerts(t, rootdir, true, true)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	cfg, err := config.LoadConfig(rootdir, flags)
	require.NoError(t, err)

	assert.Equal(t, certPath, cfg.GetString("api.pubkeypath"))
	assert.Equal(t, keyPath, cfg.GetString("api.privkeypath"))
}

func TestLoadConfigSetsOnlyFoundDefaultCertPath(t *testing.T) {
	// Only the certificate is present. The found path is set and its missing
	// pair is left empty, so the start command rejects the incomplete pair
	// rather than silently ignoring the certificate.
	rootdir := t.TempDir()
	certPath, _ := writeDefaultCerts(t, rootdir, true, false)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	cfg, err := config.LoadConfig(rootdir, flags)
	require.NoError(t, err)

	assert.Equal(t, certPath, cfg.GetString("api.pubkeypath"))
	assert.Equal(t, "", cfg.GetString("api.privkeypath"))
}

func TestLoadConfigSetsOnlyFoundDefaultKeyPath(t *testing.T) {
	// Only the key is present; symmetric to the cert-only case above.
	rootdir := t.TempDir()
	_, keyPath := writeDefaultCerts(t, rootdir, false, true)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	cfg, err := config.LoadConfig(rootdir, flags)
	require.NoError(t, err)

	assert.Equal(t, "", cfg.GetString("api.pubkeypath"))
	assert.Equal(t, keyPath, cfg.GetString("api.privkeypath"))
}

func TestLoadConfigDoesNotOverrideExplicitCertPaths(t *testing.T) {
	rootdir := t.TempDir()
	// Default certs exist, but explicitly-configured paths must take precedence
	// and auto-detection must not override them.
	writeDefaultCerts(t, rootdir, true, true)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("pubkeypath", "", "")
	flags.String("privkeypath", "", "")
	require.NoError(t, flags.Set("pubkeypath", "/custom/my.crt"))
	require.NoError(t, flags.Set("privkeypath", "/custom/my.key"))

	cfg, err := config.LoadConfig(rootdir, flags)
	require.NoError(t, err)

	assert.Equal(t, "/custom/my.crt", cfg.GetString("api.pubkeypath"))
	assert.Equal(t, "/custom/my.key", cfg.GetString("api.privkeypath"))
}
