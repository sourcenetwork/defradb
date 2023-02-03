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
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
)

func TestNewServerAndRunWithoutListener(t *testing.T) {
	ctx := context.Background()
	s := NewServer(nil, WithAddress(":0"))
	if ok := assert.NotNil(t, s); ok {
		assert.Equal(t, ErrNoListener, s.Run(ctx))
	}
}

func TestNewServerAndRunWithListenerAndInvalidPort(t *testing.T) {
	ctx := context.Background()
	s := NewServer(nil, WithAddress(":303000"))
	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(ctx))
	}
}

func TestNewServerAndRunWithListenerAndValidPort(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	s := NewServer(nil, WithAddress(":0"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.NoError(t, err)
		err = s.Run(ctx)
		assert.ErrorIs(t, http.ErrServerClosed, err)
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

func TestNewServerAndRunWithAutocertWithoutEmail(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s := NewServer(nil, WithAddress("example.com"), WithRootDir(dir), WithTLSPort(0))

	err := s.Listen(ctx)
	assert.ErrorIs(t, err, ErrNoEmail)

	s.Shutdown(context.Background())
}

func TestNewServerAndRunWithAutocert(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	dir := t.TempDir()
	s := NewServer(nil, WithAddress("example.com"), WithRootDir(dir), WithTLSPort(0), WithCAEmail("dev@defradb.net"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.NoError(t, err)
		err = s.Run(ctx)
		assert.ErrorIs(t, http.ErrServerClosed, err)
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

func TestNewServerAndRunWithSelfSignedCertAndNoKeyFiles(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	dir := t.TempDir()
	s := NewServer(nil, WithAddress("localhost:0"), WithSelfSignedCert(dir+"/server.crt", dir+"/server.key"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.Contains(t, err.Error(), "no such file or directory")
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

const pubKey = `-----BEGIN EC PARAMETERS-----
BgUrgQQAIg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDD4VK0DRBRaeieXU9JaPJfSeegGYcXaX5+gEcwGKA0UJYI46QRHIlHC
IJMOjPsrUCmgBwYFK4EEACKhZANiAAQ3ltsFK8bZZpOYiJnvwpa7Ft+b0KFsDqpu
pS0gW/SYpAncHhRuz18RQ2ycuXlSN1S/PAryRZ5PK2xORKfwpguEDEMdVwbHorZO
K44P/h3dhyNyAyf8rcRoqKXcl/K/uew=
-----END EC PRIVATE KEY-----`

const privKey = `-----BEGIN CERTIFICATE-----
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

func TestNewServerAndRunWithSelfSignedCertAndInvalidPort(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	dir := t.TempDir()
	err := os.WriteFile(dir+"/server.key", []byte(privKey), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(dir+"/server.crt", []byte(pubKey), 0644)
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(nil, WithAddress(":303000"), WithSelfSignedCert(dir+"/server.crt", dir+"/server.key"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.Contains(t, err.Error(), "invalid port")
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

func TestNewServerAndRunWithSelfSignedCert(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	dir := t.TempDir()
	err := os.WriteFile(dir+"/server.key", []byte(privKey), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(dir+"/server.crt", []byte(pubKey), 0644)
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(nil, WithAddress("localhost:0"), WithSelfSignedCert(dir+"/server.crt", dir+"/server.key"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.NoError(t, err)
		err = s.Run(ctx)
		assert.ErrorIs(t, http.ErrServerClosed, err)
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

func TestNewServerWithoutOptions(t *testing.T) {
	s := NewServer(nil)
	assert.Equal(t, "localhost:9181", s.Addr)
	assert.Equal(t, []string(nil), s.options.allowedOrigins)
}

func TestNewServerWithAddress(t *testing.T) {
	s := NewServer(nil, WithAddress("localhost:9999"))
	assert.Equal(t, "localhost:9999", s.Addr)
}

func TestNewServerWithDomainAddress(t *testing.T) {
	s := NewServer(nil, WithAddress("example.com"))
	assert.Equal(t, "example.com", s.options.domain.Value())
	assert.NotNil(t, s.options.tls)
}

func TestNewServerWithAllowedOrigins(t *testing.T) {
	s := NewServer(nil, WithAllowedOrigins("https://source.network", "https://app.source.network"))
	assert.Equal(t, []string{"https://source.network", "https://app.source.network"}, s.options.allowedOrigins)
}

func TestNewServerWithCAEmail(t *testing.T) {
	s := NewServer(nil, WithCAEmail("me@example.com"))
	assert.Equal(t, "me@example.com", s.options.tls.Value().email)
}

func TestNewServerWithPeerID(t *testing.T) {
	s := NewServer(nil, WithPeerID("12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR"))
	assert.Equal(t, "12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR", s.options.peerID)
}

func TestNewServerWithRootDir(t *testing.T) {
	dir := t.TempDir()
	s := NewServer(nil, WithRootDir(dir))
	assert.Equal(t, dir, s.options.rootDir)
}

func TestNewServerWithTLSPort(t *testing.T) {
	s := NewServer(nil, WithTLSPort(44343))
	assert.Equal(t, ":44343", s.options.tls.Value().port)
}

func TestNewServerWithSelfSignedCert(t *testing.T) {
	s := NewServer(nil, WithSelfSignedCert("pub.key", "priv.key"))
	assert.Equal(t, "pub.key", s.options.tls.Value().pubKey)
	assert.Equal(t, "priv.key", s.options.tls.Value().privKey)
	assert.NotNil(t, s.options.tls)
}

func TestNewHTTPRedirServer(t *testing.T) {
	m := &autocert.Manager{}
	s := newHTTPRedirServer(m)
	assert.Equal(t, ":80", s.Addr)
}
