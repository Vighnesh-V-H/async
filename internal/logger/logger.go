package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	Level       string
	Format      string
	ServiceName string
	Environment string
	IsProd      bool
}

func New(cfg Config) zerolog.Logger {
	var logLevel zerolog.Level
	switch cfg.Level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer
	if !cfg.IsProd || cfg.Format != "json" {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	} else {
		writer = os.Stdout
	}

	logger := zerolog.New(writer).
		Level(logLevel).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	if !cfg.IsProd {
		logger = logger.With().Stack().Logger()
	}

	return logger
}

func WithContext(logger zerolog.Logger, context map[string]any) zerolog.Logger {
	if context == nil {
		return logger
	}
	return logger.With().Fields(context).Logger()
}

func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatFieldValue: func(i any) string {
			switch v := i.(type) {
			case string:
				if len(v) > 200 {
					return v[:200] + "..."
				}
				return v
			case []byte:
				var obj interface{}
				if err := json.Unmarshal(v, &obj); err == nil {
					pretty, _ := json.MarshalIndent(obj, "", "  ")
					return "\n" + string(pretty)
				}
				return string(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("component", "database").
		Logger()
}

func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.TraceLevel:
		return 6
	case zerolog.DebugLevel:
		return 5
	case zerolog.InfoLevel:
		return 4
	case zerolog.WarnLevel:
		return 3
	case zerolog.ErrorLevel:
		return 2
	default:
		return 1
	}
}