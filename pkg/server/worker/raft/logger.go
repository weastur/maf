package raft

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

type HCZeroLogger struct {
	logger zerolog.Logger
}

func NewHCZeroLogger(logger zerolog.Logger) *HCZeroLogger {
	return &HCZeroLogger{
		logger: logger,
	}
}

func (l *HCZeroLogger) Log(level hclog.Level, format string, args ...any) {
	switch level {
	case hclog.Trace:
		l.logger.Trace().Fields(args).Msg(format)
	case hclog.Debug:
		l.logger.Debug().Fields(args).Msg(format)
	case hclog.Info:
		l.logger.Info().Fields(args).Msg(format)
	case hclog.Warn:
		l.logger.Warn().Fields(args).Msg(format)
	case hclog.Error:
		l.logger.Error().Fields(args).Msg(format)
	case hclog.NoLevel:
		l.logger.Log().Fields(args).Msg(format)
	case hclog.Off:
		// no-op
	default:
		l.logger.Fatal().Msgf("Unknown log level: %s", level)
	}
}

func (l *HCZeroLogger) Trace(format string, args ...any) {
	l.logger.Trace().Fields(args).Msg(format)
}

func (l *HCZeroLogger) Debug(format string, args ...any) {
	l.logger.Debug().Fields(args).Msg(format)
}

func (l *HCZeroLogger) Info(format string, args ...any) {
	l.logger.Info().Fields(args).Msg(format)
}

func (l *HCZeroLogger) Warn(format string, args ...any) {
	l.logger.Warn().Fields(args).Msg(format)
}

func (l *HCZeroLogger) Error(format string, args ...any) {
	l.logger.Error().Fields(args).Msg(format)
}

func (l *HCZeroLogger) IsTrace() bool {
	return l.logger.GetLevel() == zerolog.TraceLevel
}

func (l *HCZeroLogger) IsDebug() bool {
	return l.logger.GetLevel() == zerolog.DebugLevel
}

func (l *HCZeroLogger) IsInfo() bool {
	return l.logger.GetLevel() == zerolog.InfoLevel
}

func (l *HCZeroLogger) IsWarn() bool {
	return l.logger.GetLevel() == zerolog.WarnLevel
}

func (l *HCZeroLogger) IsError() bool {
	return l.logger.GetLevel() == zerolog.ErrorLevel
}

func (l *HCZeroLogger) ImpliedArgs() []any {
	return nil
}

func (l *HCZeroLogger) With(args ...any) hclog.Logger {
	return &HCZeroLogger{l.logger.With().Fields(args).Logger()}
}

func (l *HCZeroLogger) Name() string {
	return ""
}

func (l *HCZeroLogger) Named(name string) hclog.Logger {
	return &HCZeroLogger{l.logger.With().Str("name", name).Logger()}
}

func (l *HCZeroLogger) ResetNamed(name string) hclog.Logger {
	return &HCZeroLogger{l.logger.With().Str("name", name).Logger()}
}

func (l *HCZeroLogger) SetLevel(level hclog.Level) {
	switch level {
	case hclog.Trace:
		l.logger = l.logger.Level(zerolog.TraceLevel)
	case hclog.Debug:
		l.logger = l.logger.Level(zerolog.DebugLevel)
	case hclog.Info:
		l.logger = l.logger.Level(zerolog.InfoLevel)
	case hclog.Warn:
		l.logger = l.logger.Level(zerolog.WarnLevel)
	case hclog.Error:
		l.logger = l.logger.Level(zerolog.ErrorLevel)
	case hclog.Off:
		l.logger = l.logger.Level(zerolog.Disabled)
	case hclog.NoLevel:
		l.logger = l.logger.Level(zerolog.NoLevel)
	default:
		l.logger.Fatal().Msgf("Unknown log level: %s", level)
	}
}

func (l *HCZeroLogger) GetLevel() hclog.Level {
	switch l.logger.GetLevel() {
	case zerolog.TraceLevel:
		return hclog.Trace
	case zerolog.DebugLevel:
		return hclog.Debug
	case zerolog.InfoLevel:
		return hclog.Info
	case zerolog.WarnLevel:
		return hclog.Warn
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		return hclog.Error
	case zerolog.Disabled:
		return hclog.Off
	case zerolog.NoLevel:
		return hclog.NoLevel
	default:
		l.logger.Fatal().Msgf("Unknown log level: %s", l.logger.GetLevel())

		return hclog.NoLevel
	}
}

func (l *HCZeroLogger) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(l.logger, "", 0)
}

func (l *HCZeroLogger) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return l.logger
}
