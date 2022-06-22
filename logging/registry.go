// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package logging

import (
	"sync"
)

var configMutex sync.RWMutex
var cachedConfig Config

var registryMutex sync.Mutex
var registry map[string]Logger = map[string]Logger{}

func register(name string, logger Logger) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	registry[name] = logger
}

func setConfig(newConfig Config) Config {
	configMutex.Lock()
	defer configMutex.Unlock()

	cachedConfig = cachedConfig.with(newConfig)
	return cachedConfig
}

func updateLoggers(config Config) {
	for loggerName, logger := range registry {
		newLoggerConfig := config.forLogger(loggerName)
		logger.ApplyConfig(newLoggerConfig)
	}
}
