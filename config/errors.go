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
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToWriteFile           string = "failed to write file"
	errFailedToRemoveConfigFile    string = "failed to remove config file"
	errCannotBeHomeDir                    = "path cannot be just ~ (home directory)"
	errUnableToExpandHomeDir              = "unable to expand home directory"
	errNoDatabaseURLProvided              = "no database URL provided"
	errLoggingConfigNotObtained           = "could not get logging config"
	errFailedToValidateConfig             = "failed to validate config"
	errOverrideConfigConvertFailed        = "invalid override config"
	errConfigToJSONFailed                 = "failed to marshal Config to JSON"
	errInvalidDatabaseURL                 = "invalid database URL"
	errInvalidRPCTimeout                  = "invalid RPC timeout"
	errInvalidRPCMaxConnectionIdle        = "invalid RPC MaxConnectionIdle"
	errInvalidP2PAddress                  = "invalid P2P address"
	errInvalidRPCAddress                  = "invalid RPC address"
	errInvalidBootstrapPeers              = "invalid bootstrap peers"
	errInvalidLogLevel                    = "invalid log level"
	errInvalidStoreType                   = "invalid store type"
	errInvalidLogFormat                   = "invalid log format"
	errInvalidNamedLoggerName             = "invalid named logger name"
	errConfigTemplateFailed               = "could not process config template"
)

var (
	ErrFailedToWriteFile           = errors.New(errFailedToWriteFile)
	ErrFailedToRemoveConfigFile    = errors.New(errFailedToRemoveConfigFile)
	ErrPathCannotBeHomeDir         = errors.New(errCannotBeHomeDir)
	ErrUnableToExpandHomeDir       = errors.New(errUnableToExpandHomeDir)
	ErrNoDatabaseURLProvided       = errors.New(errNoDatabaseURLProvided)
	ErrInvalidDatabaseURL          = errors.New(errInvalidDatabaseURL)
	ErrLoggingConfigNotObtained    = errors.New(errLoggingConfigNotObtained)
	ErrFailedToValidateConfig      = errors.New(errFailedToValidateConfig)
	ErrInvalidRPCTimeout           = errors.New(errInvalidRPCTimeout)
	ErrInvalidRPCMaxConnectionIdle = errors.New(errInvalidRPCMaxConnectionIdle)
	ErrInvalidP2PAddress           = errors.New(errInvalidP2PAddress)
	ErrInvalidRPCAddress           = errors.New(errInvalidRPCAddress)
	ErrInvalidBootstrapPeers       = errors.New(errInvalidBootstrapPeers)
	ErrInvalidLogLevel             = errors.New(errInvalidLogLevel)
	ErrInvalidStoreType            = errors.New(errInvalidStoreType)
	ErrOverrideConfigConvertFailed = errors.New(errOverrideConfigConvertFailed)
	ErrInvalidLogFormat            = errors.New(errInvalidLogFormat)
	ErrConfigToJSONFailed          = errors.New(errConfigToJSONFailed)
	ErrInvalidNamedLoggerName      = errors.New(errInvalidNamedLoggerName)
	ErrConfigTemplateFailed        = errors.New(errConfigTemplateFailed)
)

func NewErrFailedToWriteFile(inner error, path string) error {
	return errors.Wrap(errFailedToWriteFile, inner, errors.NewKV("path", path))
}

func NewErrFailedToRemoveConfigFile(inner error) error {
	return errors.Wrap(errFailedToRemoveConfigFile, inner)
}

func NewErrPathCannotBeHomeDir(inner error) error {
	return errors.Wrap(errCannotBeHomeDir, inner)
}

func NewErrUnableToExpandHomeDir(inner error) error {
	return errors.Wrap(errUnableToExpandHomeDir, inner)
}

func NewErrNoDatabaseURLProvided(inner error) error {
	return errors.Wrap(errNoDatabaseURLProvided, inner)
}

func NewErrInvalidDatabaseURL(inner error) error {
	return errors.Wrap(errInvalidDatabaseURL, inner)
}

func NewErrLoggingConfigNotObtained(inner error) error {
	return errors.Wrap(errLoggingConfigNotObtained, inner)
}

func NewErrFailedToValidateConfig(inner error) error {
	return errors.Wrap(errFailedToValidateConfig, inner)
}

func NewErrInvalidRPCTimeout(inner error, timeout string) error {
	return errors.Wrap(errInvalidRPCTimeout, inner, errors.NewKV("timeout", timeout))
}

func NewErrInvalidRPCMaxConnectionIdle(inner error, timeout string) error {
	return errors.Wrap(errInvalidRPCMaxConnectionIdle, inner, errors.NewKV("timeout", timeout))
}

func NewErrInvalidP2PAddress(inner error, address string) error {
	return errors.Wrap(errInvalidP2PAddress, inner, errors.NewKV("address", address))
}

func NewErrInvalidRPCAddress(inner error, address string) error {
	return errors.Wrap(errInvalidRPCAddress, inner, errors.NewKV("address", address))
}

func NewErrInvalidBootstrapPeers(inner error, peers string) error {
	return errors.Wrap(errInvalidBootstrapPeers, inner, errors.NewKV("peers", peers))
}

func NewErrInvalidLogLevel(level string) error {
	return errors.New(errInvalidLogLevel, errors.NewKV("level", level))
}

func NewErrInvalidDatastoreType(storeType string) error {
	return errors.New(errInvalidStoreType, errors.NewKV("store_type", storeType))
}

func NewErrOverrideConfigConvertFailed(inner error, name string) error {
	return errors.Wrap(errOverrideConfigConvertFailed, inner, errors.NewKV("name", name))
}

func NewErrInvalidLogFormat(format string) error {
	return errors.New(errInvalidLogFormat, errors.NewKV("format", format))
}

func NewErrConfigToJSONFailed(inner error) error {
	return errors.Wrap(errConfigToJSONFailed, inner)
}

func NewErrInvalidNamedLoggerName(name string) error {
	return errors.New(errInvalidNamedLoggerName, errors.NewKV("name", name))
}

func NewErrConfigTemplateFailed(inner error) error {
	return errors.Wrap(errConfigTemplateFailed, inner)
}
