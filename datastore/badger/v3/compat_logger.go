// Code copy-pasted from https://github.com/ipfs/go-ds-badger2/blob/master/compat_logger.go

package badger

import "go.uber.org/zap"

type compatLogger struct {
	zap.SugaredLogger
	skipLogger zap.SugaredLogger
}

// Warning is for compatibility
// Deprecated: use Warn(args ...any) instead
func (logger *compatLogger) Warning(args ...any) {
	logger.skipLogger.Warn(args...)
}

// Warningf is for compatibility
// Deprecated: use Warnf(format string, args ...any) instead
func (logger *compatLogger) Warningf(format string, args ...any) {
	logger.skipLogger.Warnf(format, args...)
}
