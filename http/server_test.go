// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tlsKey is the TLS private key in PEM format
const tlsKey = `-----BEGIN EC PARAMETERS-----
BgUrgQQAIg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDD4VK0DRBRaeieXU9JaPJfSeegGYcXaX5+gEcwGKA0UJYI46QRHIlHC
IJMOjPsrUCmgBwYFK4EEACKhZANiAAQ3ltsFK8bZZpOYiJnvwpa7Ft+b0KFsDqpu
pS0gW/SYpAncHhRuz18RQ2ycuXlSN1S/PAryRZ5PK2xORKfwpguEDEMdVwbHorZO
K44P/h3dhyNyAyf8rcRoqKXcl/K/uew=
-----END EC PRIVATE KEY-----`

// tlsKey is the TLS certificate in PEM format
const tlsCert = `-----BEGIN CERTIFICATE-----
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

// insecureClient is an http client that trusts all tls certificates
var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// testHandler returns an empty body and 200 status code
var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestServerServeWithNoListener(t *testing.T) {
	srv, err := NewServer(testHandler)
	require.NoError(t, err)

	err = srv.Serve()
	require.ErrorIs(t, err, ErrNoListener)
}

func TestServerServeWithTLSAndNoListener(t *testing.T) {
	certPath, keyPath := writeTestCerts(t)
	srv, err := NewServer(testHandler, WithTLSCertPath(certPath), WithTLSKeyPath(keyPath))
	require.NoError(t, err)

	err = srv.Serve()
	require.ErrorIs(t, err, ErrNoListener)
}

func TestServerListenAndServeWithInvalidAddress(t *testing.T) {
	srv, err := NewServer(testHandler, WithAddress("invalid"))
	require.NoError(t, err)

	err = srv.SetListener()
	require.ErrorContains(t, err, "address invalid")
}

func TestServerListenAndServeWithTLSAndInvalidAddress(t *testing.T) {
	certPath, keyPath := writeTestCerts(t)
	srv, err := NewServer(testHandler, WithAddress("invalid"), WithTLSCertPath(certPath), WithTLSKeyPath(keyPath))
	require.NoError(t, err)

	err = srv.SetListener()
	require.ErrorContains(t, err, "address invalid")
}

func TestServerListenAndServeWithTLSAndInvalidCerts(t *testing.T) {
	srv, err := NewServer(
		testHandler,
		WithAddress("invalid"),
		WithTLSCertPath("invalid.crt"),
		WithTLSKeyPath("invalid.key"),
		WithAddress("127.0.0.1:30001"),
	)
	require.NoError(t, err)

	err = srv.SetListener()
	require.NoError(t, err)
	err = srv.Serve()
	require.ErrorContains(t, err, "no such file or directory")
	err = srv.listener.Close()
	require.NoError(t, err)
}

func TestServerListenAndServeWithAddress(t *testing.T) {
	srv, err := NewServer(testHandler, WithAddress("127.0.0.1:30001"))
	require.NoError(t, err)

	go func() {
		err := srv.SetListener()
		require.NoError(t, err)
		err = srv.Serve()
		require.ErrorIs(t, http.ErrServerClosed, err)
	}()

	// wait for server to start
	<-time.After(time.Second * 1)

	res, err := http.Get("http://127.0.0.1:30001")
	require.NoError(t, err)

	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	err = srv.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestServerListenAndServeWithTLS(t *testing.T) {
	certPath, keyPath := writeTestCerts(t)
	srv, err := NewServer(testHandler, WithAddress("127.0.0.1:8443"), WithTLSCertPath(certPath), WithTLSKeyPath(keyPath))
	require.NoError(t, err)

	go func() {
		err := srv.SetListener()
		require.NoError(t, err)
		err = srv.Serve()
		require.ErrorIs(t, http.ErrServerClosed, err)
	}()

	// wait for server to start
	<-time.After(time.Second * 1)

	res, err := insecureClient.Get("https://127.0.0.1:8443")
	require.NoError(t, err)

	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)

	err = srv.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestServerListenAndServeWithAllowedOrigins(t *testing.T) {
	srv, err := NewServer(testHandler, WithAllowedOrigins("localhost"), WithAddress("127.0.0.1:30001"))
	require.NoError(t, err)

	go func() {
		err := srv.SetListener()
		require.NoError(t, err)
		err = srv.Serve()
		require.ErrorIs(t, http.ErrServerClosed, err)
	}()

	// wait for server to start
	<-time.After(time.Second * 1)

	req, err := http.NewRequest(http.MethodOptions, "http://127.0.0.1:30001", nil)
	require.NoError(t, err)
	req.Header.Add("origin", "localhost")
	req.Header.Add("Access-Control-Request-Method", "POST")
	req.Header.Add("Access-Control-Request-Headers", "Authorization, Content-Type")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "localhost", res.Header.Get("Access-Control-Allow-Origin"))

	err = srv.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestServerWithReadTimeout(t *testing.T) {
	srv, err := NewServer(testHandler, WithReadTimeout(time.Second))
	require.NoError(t, err)
	assert.Equal(t, time.Second, srv.server.ReadTimeout)
}

func TestServerWithWriteTimeout(t *testing.T) {
	srv, err := NewServer(testHandler, WithWriteTimeout(time.Second))
	require.NoError(t, err)
	assert.Equal(t, time.Second, srv.server.WriteTimeout)
}

func TestServerWithIdleTimeout(t *testing.T) {
	srv, err := NewServer(testHandler, WithIdleTimeout(time.Second))
	require.NoError(t, err)
	assert.Equal(t, time.Second, srv.server.IdleTimeout)
}

func writeTestCerts(t *testing.T) (string, string) {
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "cert.pub")
	keyPath := filepath.Join(tempDir, "cert.key")

	err := os.WriteFile(certPath, []byte(tlsCert), 0644)
	require.NoError(t, err)

	err = os.WriteFile(keyPath, []byte(tlsKey), 0644)
	require.NoError(t, err)

	return certPath, keyPath
}
