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
	errCannotBeHomeDir             string = "path cannot be just ~ (home directory)"
	errUnableToExpandHomeDir       string = "unable to expand home directory"
	errNoDatabaseURLProvided       string = "no database URL provided"
	errLoggingConfigNotObtained    string = "could not get logging config"
	errFailedToValidateConfig      string = "failed to validate config"
	errOverrideConfigConvertFailed string = "invalid override config"
	errConfigToJSONFailed          string = "failed to marshal Config to JSON"
	errInvalidDatabaseURL          string = "invalid database URL"
	errInvalidRPCTimeout           string = "invalid RPC timeout"
	errInvalidRPCMaxConnectionIdle string = "invalid RPC MaxConnectionIdle"
	errInvalidP2PAddress           string = "invalid P2P address"
	errInvalidRPCAddress           string = "invalid RPC address"
	errInvalidBootstrapPeers       string = "invalid bootstrap peers"
	errInvalidLogLevel             string = "invalid log level"
	errInvalidDatastoreType        string = "invalid store type"
	errInvalidLogFormat            string = "invalid log format"
	errInvalidNamedLoggerName      string = "invalid named logger name"
	errInvalidLoggerConfig         string = "invalid logger config"
	errConfigTemplateFailed        string = "could not process config template"
	errCouldNotObtainLoggerConfig  string = "could not get named logger config"
	errNotProvidedAsKV             string = "logging config parameter was not provided as <key>=<value> pair"
	errLoggerNameEmpty             string = "logger name cannot be empty"
	errCouldNotParseType           string = "could not parse type"
	errUnknownLoggerParameter      string = "unknown logger parameter"
	errInvalidLoggerName           string = "invalid logger name"
	errDuplicateLoggerName         string = "duplicate logger name"
	errReadingConfigFile           string = "failed to read config"
	errLoadingConfig               string = "failed to load config"
	errUnableToParseByteSize       string = "unable to parse byte size"
	errInvalidDatastorePath        string = "invalid datastore path"
	errMissingPortNumber           string = "missing port number"
	errNoPortWithDomain            string = "cannot provide port with domain name"
	errInvalidRootDir              string = "invalid root directory"
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
	ErrInvalidDatastoreType        = errors.New(errInvalidDatastoreType)
	ErrOverrideConfigConvertFailed = errors.New(errOverrideConfigConvertFailed)
	ErrInvalidLogFormat            = errors.New(errInvalidLogFormat)
	ErrConfigToJSONFailed          = errors.New(errConfigToJSONFailed)
	ErrInvalidNamedLoggerName      = errors.New(errInvalidNamedLoggerName)
	ErrConfigTemplateFailed        = errors.New(errConfigTemplateFailed)
	ErrCouldNotObtainLoggerConfig  = errors.New(errCouldNotObtainLoggerConfig)
	ErrNotProvidedAsKV             = errors.New(errNotProvidedAsKV)
	ErrLoggerNameEmpty             = errors.New(errLoggerNameEmpty)
	ErrCouldNotParseType           = errors.New(errCouldNotParseType)
	ErrUnknownLoggerParameter      = errors.New(errUnknownLoggerParameter)
	ErrInvalidLoggerName           = errors.New(errInvalidLoggerName)
	ErrDuplicateLoggerName         = errors.New(errDuplicateLoggerName)
	ErrReadingConfigFile           = errors.New(errReadingConfigFile)
	ErrLoadingConfig               = errors.New(errLoadingConfig)
	ErrUnableToParseByteSize       = errors.New(errUnableToParseByteSize)
	ErrInvalidLoggerConfig         = errors.New(errInvalidLoggerConfig)
	ErrorInvalidDatastorePath      = errors.New(errInvalidDatastorePath)
	ErrMissingPortNumber           = errors.New(errMissingPortNumber)
	ErrNoPortWithDomain            = errors.New(errNoPortWithDomain)
	ErrorInvalidRootDir            = errors.New(errInvalidRootDir)
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
	return errors.New(errInvalidDatastoreType, errors.NewKV("store_type", storeType))
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

func NewErrCouldNotObtainLoggerConfig(inner error, name string) error {
	return errors.Wrap(errCouldNotObtainLoggerConfig, inner, errors.NewKV("name", name))
}

func NewErrNotProvidedAsKV(kv string) error {
	return errors.New(errNotProvidedAsKV, errors.NewKV("KV", kv))
}

func NewErrCouldNotParseType(inner error, name string) error {
	return errors.Wrap(errCouldNotParseType, inner, errors.NewKV("name", name))
}

func NewErrUnknownLoggerParameter(name string) error {
	return errors.New(errUnknownLoggerParameter, errors.NewKV("param", name))
}

func NewErrInvalidLoggerName(name string) error {
	return errors.New(errInvalidLoggerName, errors.NewKV("name", name))
}

func NewErrDuplicateLoggerName(name string) error {
	return errors.New(errDuplicateLoggerName, errors.NewKV("name", name))
}

func NewErrReadingConfigFile(inner error) error {
	return errors.Wrap(errReadingConfigFile, inner)
}

func NewErrLoadingConfig(inner error) error {
	return errors.Wrap(errLoadingConfig, inner)
}

func NewErrUnableToParseByteSize(err error) error {
	return errors.Wrap(errUnableToParseByteSize, err)
}

func NewErrLoggerConfig(s string) error {
	return errors.New(errInvalidLoggerConfig, errors.NewKV("explanation", s))
}

func NewErrInvalidDatastorePath(path string) error {
	return errors.New(errInvalidDatastorePath, errors.NewKV("path", path))
}

func NewErrInvalidRootDir(path string) error {
	return errors.New(errInvalidRootDir, errors.NewKV("path", path))
}
