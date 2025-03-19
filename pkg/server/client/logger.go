package client

import "github.com/rs/zerolog"

type RestyLogger struct {
	logger zerolog.Logger
}

func NewRestyLogger(logger zerolog.Logger) *RestyLogger {
	return &RestyLogger{
		logger: logger,
	}
}

func (r *RestyLogger) Errorf(format string, v ...any) {
	r.logger.Error().Msgf(format, v...)
}

func (r *RestyLogger) Warnf(format string, v ...any) {
	r.logger.Warn().Msgf(format, v...)
}

func (r *RestyLogger) Debugf(format string, v ...any) {
	r.logger.Debug().Msgf(format, v...)
}
