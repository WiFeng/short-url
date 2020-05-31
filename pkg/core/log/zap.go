package log

import (
	"github.com/WiFeng/short-url/pkg/core/config"
	"go.uber.org/zap"
)

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
