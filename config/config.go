package config

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/viper"
)

// configName is the config file name
const configName = "config.yaml"

//go:embed config.yaml
var defaultConfig []byte

// relativePathKeys are config keys that will be made relative to the rootdir
var relativePathKeys = []string{
	"datastore.badger.path",
	"api.pubkeypath",
	"api.privkeypath",
}

// WriteDefaultConfig writes the default config file to the given rootdir.
func WriteDefaultConfig(rootdir string) error {
	if err := os.MkdirAll(rootdir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(rootdir, configName), defaultConfig, 0666)
}

// LoadConfig returns a new config with values from the config in the given rootdir.
func LoadConfig(rootdir string) (*viper.Viper, error) {
	cfg := viper.New()

	cfg.SetEnvPrefix("DEFRA")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cfg.SetConfigName(configName)
	cfg.SetConfigType("yaml")

	cfg.AddConfigPath(rootdir)
	cfg.AutomaticEnv()

	// load defaults first then merge persisted config
	err := cfg.MergeConfig(bytes.NewBuffer(defaultConfig))
	if err != nil {
		return nil, err
	}
	err = cfg.MergeInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		return nil, err
	}

	// make paths relative to the rootdir
	for _, key := range relativePathKeys {
		path := cfg.GetString(key)
		if !filepath.IsAbs(path) {
			cfg.Set(key, filepath.Join(rootdir, path))
		}
	}

	var level int8
	switch value := cfg.GetString("log.level"); value {
	case "debug":
		level = logging.Debug
	case "info":
		level = logging.Info
	case "error":
		level = logging.Error
	case "fatal":
		level = logging.Fatal
	default:
		level = logging.Info
	}

	var format logging.EncoderFormat
	switch value := cfg.GetString("log.format"); value {
	case "json":
		format = logging.JSON
	case "csv":
		format = logging.CSV
	default:
		format = logging.CSV
	}

	logging.SetConfig(logging.Config{
		Level:            logging.NewLogLevelOption(level),
		EnableStackTrace: logging.NewEnableStackTraceOption(cfg.GetBool("log.stacktrace")),
		DisableColor:     logging.NewDisableColorOption(cfg.GetBool("log.nocolor")),
		EncoderFormat:    logging.NewEncoderFormatOption(format),
		OutputPaths:      []string{cfg.GetString("log.output")},
		EnableCaller:     logging.NewEnableCallerOption(cfg.GetBool("log.caller")),
	})

	return cfg, nil
}
