package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New(level string) *Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	log := zerolog.New(os.Stdout).
		Level(parseLevel(level)).
		With().
		Timestamp().
		Caller().
		Logger()
	return &Logger{Logger: log}
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) WithRequestID(reqID string) *Logger {
	return &Logger{
		Logger: l.With().Str("request_id", reqID).Logger(),
	}
}

func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.With().Str("component", component).Logger(),
	}
}

func (l *Logger) Info(msg string, keyvals ...any) {
	l.Logger.Info().Fields(keyvals).Msg(msg)
}

func (l *Logger) Error(msg string, keyvals ...any) {
	l.Logger.Error().Fields(keyvals).Msg(msg)
}

func (l *Logger) Debug(msg string, keyvals ...any) {
	l.Logger.Debug().Fields(keyvals).Msg(msg)
}

func (l *Logger) Warn(msg string, keyvals ...any) {
	l.Logger.Warn().Fields(keyvals).Msg(msg)
}