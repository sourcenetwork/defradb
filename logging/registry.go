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
var registry = map[string][]Logger{
	"reprovider.simple": {GetGoLogger("reprovider.simple")},
	"badger":            {GetGoLoggerV2("badger")},
}

func register(name string, logger Logger) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	loggers, exists := registry[name]
	if !exists {
		loggers = []Logger{}
	}
	loggers = append(loggers, logger)
	registry[name] = loggers
}

func setConfig(newConfig Config) Config {
	configMutex.Lock()
	defer configMutex.Unlock()

	cachedConfig = cachedConfig.with(newConfig)
	return cachedConfig
}

func updateLoggers(config Config) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	for loggerName, loggers := range registry {
		newLoggerConfig := config.forLogger(loggerName)

		for _, logger := range loggers {
			logger.ApplyConfig(newLoggerConfig)
		}
	}
}
