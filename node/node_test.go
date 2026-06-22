// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// testTLSKey and testTLSCert are a self-signed EC key/cert pair used only in
// tests to exercise TLS startup without hitting the system CA pool.
const testTLSKey = `-----BEGIN EC PARAMETERS-----
BgUrgQQAIg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDD4VK0DRBRaeieXU9JaPJfSeegGYcXaX5+gEcwGKA0UJYI46QRHIlHC
IJMOjPsrUCmgBwYFK4EEACKhZANiAAQ3ltsFK8bZZpOYiJnvwpa7Ft+b0KFsDqpu
pS0gW/SYpAncHhRuz18RQ2ycuXlSN1S/PAryRZ5PK2xORKfwpguEDEMdVwbHorZO
K44P/h3dhyNyAyf8rcRoqKXcl/K/uew=
-----END EC PRIVATE KEY-----`

const testTLSCert = `-----BEGIN CERTIFICATE-----
MIICQDCCAcUCCQDpMnN1gQ4fGTAKBggqhkjOPQQDAjCBiDELMAkGA1UEBhMCY2Ex
DzANBgNVBAgMBlF1ZWJlYzEQMA4GA1UEBwwHQ2hlbHNlYTEPMA0GA1UECgwGU291
cmNlMRAwDgYDVQQLDAdEZWZyYURCMQ8wDQYDVQQDDAZzb3VyY2UxIjAgBgkqhkiG
9w0BCQEWE2V4YW1wbGVAZXhhbXBsZS5jb20wHhcNMjIxMDA2MTgyMjE1WhcNMjMx
MDA2MTgyMjE1WjCBiDELMAkGA1UEBhMCY2ExDzANBgNVBAgMBlF1ZWJlYzEQMA4G
A1UEBwwHQ2hlbHNlYTEPMA0GA1UECgwGU291cmNlMRAwDgYDVQQLDAdEZWZyYURC
MQ8wDQYDVQQDDAZzb3VyY2UxIjAgBgkqhkiG9w0BCQEWE2V4YW1wbGVAZXhhbXBs
ZS5jb20wdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAQ3ltsFK8bZZpOYiJnvwpa7Ft+b
0KFsDqpupS0gW/SYpAncHhRuz18RQ2ycuXlSN1S/PAryRZ5PK2xORKfwpguEDEMd
VwbHorZOK44P/h3dhyNyAyf8rcRoqKXcl/K/uewwCgYIKoZIzj0EAwIDaQAwZgIx
AIfNQeo8syOb94ojF40jY+fY1ZBSbNNK6UUbFquwDMVEoSyXRJHHEU12NUKCVTUH
kgIxAKaEGC+lqp0aaN+yubYLRiTDxOlNpyiHox3nZiL4bG/CCdPDvbX63QcdI2yq
XPKczg==
-----END CERTIFICATE-----`

func writeTestCerts(t *testing.T) (certPath, keyPath string) {
	t.Helper()
	dir := t.TempDir()
	certPath = filepath.Join(dir, "cert.pub")
	keyPath = filepath.Join(dir, "cert.key")
	require.NoError(t, os.WriteFile(certPath, []byte(testTLSCert), 0644))
	require.NoError(t, os.WriteFile(keyPath, []byte(testTLSKey), 0644))
	return certPath, keyPath
}

func TestStartAPIWithTLS(t *testing.T) {
	certPath, keyPath := writeTestCerts(t)
	ctx := context.Background()

	n, err := New(ctx,
		options.Node().
			SetDisableP2P(true).
			Store().SetPath(t.TempDir()).
			Node().
			HTTP().SetAddress("127.0.0.1:0").SetCertPath(certPath).SetKeyPath(keyPath).
			Node(),
	)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)
	defer n.Close(ctx) //nolint:errcheck

	assert.True(t, strings.HasPrefix(n.APIURL, "https://"), "expected https:// APIURL, got %s", n.APIURL)
}

func TestPurgeAndRestartWithDevModeDisabled(t *testing.T) {
	ctx := context.Background()

	n, err := New(ctx,
		options.Node().
			SetDisableAPI(true).
			SetDisableP2P(true).
			Store().SetPath(t.TempDir()).
			Node(),
	)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.ErrorIs(t, err, client.ErrOperationRequiresDeveloperMode)
}

func TestPurgeAndRestartWithDevModeEnabled(t *testing.T) {
	ctx := context.Background()

	n, err := New(ctx,
		options.Node().
			SetDisableAPI(true).
			SetDisableP2P(true).
			SetEnableDevelopment(true).
			Store().SetPath(t.TempDir()).
			Node(),
	)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	_, err = n.DB.AddCollection(ctx, "type User { name: String }")
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.NoError(t, err)

	collections, err := n.DB.GetCollections(ctx)
	require.NoError(t, err)

	assert.Len(t, collections, 0)
}
