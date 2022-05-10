// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	badgerds "github.com/ipfs/go-ds-badger2"
	"github.com/sourcenetwork/defradb/logging"
)

type Config struct {
	Database Options
	Net      NetOptions
	Logging  BaseLoggerOptions
}

type Options struct {
	Address string
	Store   string
	Memory  MemoryOptions
	Badger  BadgerOptions
}

// BadgerOptions for the badger instance of the backing datastore
type BadgerOptions struct {
	Path string
	*badgerds.Options
}

// MemoryOptions for the memory instance of the backing datastore
type MemoryOptions struct {
	Size uint64
}

type NetOptions struct {
	P2PAddress  string
	P2PDisabled bool
	TCPAddress  string
}

type BaseLoggerOptions struct {
	Level            *string
	EnableStackTrace *bool
	EncoderFormat    *string
	OutputPaths      *[]string
	NamedOptions     *[]NamedLoggerOptions
}

type NamedLoggerOptions struct {
	Name             string
	Level            *string
	EnableStackTrace *bool
	EncoderFormat    *string
	OutputPaths      *[]string
}

func (o BaseLoggerOptions) toLogConfig() logging.Config {
	var level logging.LogLevelOption
	if o.Level != nil {
		level = getLogLevelFromString(*o.Level)
	}

	var enableStackTrace logging.EnableStackTraceOption
	if o.EnableStackTrace != nil {
		enableStackTrace = logging.NewEnableStackTraceOption(*o.EnableStackTrace)
	}

	var encoderFormat logging.EncoderFormatOption
	if o.EncoderFormat != nil {
		switch *o.EncoderFormat {
		case "json":
			encoderFormat = logging.NewEncoderFormatOption(logging.JSON)
		case "csv":
			encoderFormat = logging.NewEncoderFormatOption(logging.CSV)
		}
	}

	var outputPaths []string
	if o.OutputPaths != nil {
		outputPaths = *o.OutputPaths
	}

	return logging.Config{
		Level:            level,
		EnableStackTrace: enableStackTrace,
		EncoderFormat:    encoderFormat,
		OutputPaths:      outputPaths,
	}
}

func getLogLevelFromString(logLevel string) logging.LogLevelOption {
	switch logLevel {
	case "debug":
		return logging.NewLogLevelOption(logging.Debug)
	case "info":
		return logging.NewLogLevelOption(logging.Info)
	case "warn":
		return logging.NewLogLevelOption(logging.Warn)
	case "error":
		return logging.NewLogLevelOption(logging.Error)
	case "fatal":
		return logging.NewLogLevelOption(logging.Fatal)
	default:
		return logging.LogLevelOption{}
	}
}

var (
	defaultConfig = Config{
		Database: Options{
			Address: "localhost:9181",
			Store:   "badger",
			Badger: BadgerOptions{
				Path: "$HOME/.defradb/data",
			},
		},
		Net: NetOptions{
			P2PAddress: "/ip4/0.0.0.0/tcp/9171",
			TCPAddress: "/ip4/0.0.0.0/tcp/9161",
		},
	}
)
