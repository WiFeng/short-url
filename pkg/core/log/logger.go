package log

import (
	"context"

	"github.com/WiFeng/short-url/pkg/core/config"
	"go.uber.org/zap"
)

// Logger interface
type Logger interface {
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	DPanicw(msg string, keysAndValues ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Sync() error
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	With2(args ...interface{}) Logger

	Log(keyvals ...interface{}) error
}

type logger struct {
	*zap.SugaredLogger
}

// NewLogger new Logger
func NewLogger(conf *config.Config) (Logger, error) {
	zapConf := config.NewZapConfig(conf)
	zapLogger, err := zapConf.Build()
	if err != nil {
		return nil, err
	}
	logger := logger{
		zapLogger.Sugar(),
	}
	return logger, nil
}

func (l logger) Log(keyvals ...interface{}) error {
	l.Info(keyvals...)
	return nil
}

func (l logger) With2(args ...interface{}) Logger {
	sl := l.With(args...)

	logger := logger{
		sl,
	}

	return logger
}

// default logger
var defaultLogger Logger

// GetDefaultLogger return default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// SetDefaultLogger set default logger
func SetDefaultLogger(logg Logger) {
	defaultLogger = logg
}

// Infow function
func Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logg := LoggerFromContext(ctx)
	logg.Infow(msg, keysAndValues...)
}
