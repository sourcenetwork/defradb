// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errUnknownKMSType              string = "unknown KMS type"
	errNoPeerSuppliedEncryptionKey string = "no peer supplied the encryption key"

	errPeerErrorKeyReply           string = "encryption key peer reply carried error"
	errEncryptionKeyUnmarshal      string = "unmarshal encryption key response"
	errReplyLinksAndBlocksMismatch string = "encryption key peer reply had mismatched links and blocks"
	errDecryptEncryptionKey        string = "decrypting encryption key"
	errDecodingEncryptionKey       string = "decoding encryption key"
	errEncryptionKeyCIDMismatch    string = "key CID mismatch"
	errEncryptionKeyStore          string = "store encryption key"
	errKeyCIDGeneration            string = "key cid generation"
)

var (
	ErrUnknownKMSType              = errors.New(errUnknownKMSType)
	ErrNoPeerSuppliedEncryptionKey = errors.New(errNoPeerSuppliedEncryptionKey)

	ErrPeerErrorKeyReply           = errors.New(errPeerErrorKeyReply)
	ErrEncryptionKeyUnmarshal      = errors.New(errEncryptionKeyUnmarshal)
	ErrReplyLinksAndBlocksMismatch = errors.New(errReplyLinksAndBlocksMismatch)
	ErrDecryptEncryptionKey        = errors.New(errDecryptEncryptionKey)
	ErrDecodingEncryptionKey       = errors.New(errDecodingEncryptionKey)
	ErrEncryptionKeyCIDMismatch    = errors.New(errEncryptionKeyCIDMismatch)
	ErrEncryptionKeyStore          = errors.New(errEncryptionKeyStore)
	ErrKeyCIDGeneration            = errors.New(errKeyCIDGeneration)
)

func NewErrUnknownKMSType(t ServiceType) error {
	return errors.New(errUnknownKMSType, errors.NewKV("Type", t))
}

func NewErrReplyLinksAndBlocksMismatch(linkCount, blockCount int) error {
	return errors.New(errReplyLinksAndBlocksMismatch,
		errors.NewKV("LinkCount", linkCount),
		errors.NewKV("BlockCount", blockCount),
	)
}
