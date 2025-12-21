package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init(stage string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if stage == "dev" {
		Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		Logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Str("stage", stage).
			Logger()
	}
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func WithField(key string, value interface{}) zerolog.Logger {
	return Logger.With().Interface(key, value).Logger()
}
