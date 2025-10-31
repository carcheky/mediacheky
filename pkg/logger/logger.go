package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level string) *Logger {
	return NewWithFile(level, "")
}

func NewWithFile(level string, logFilePath string) *Logger {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create encoder config for console (with colors)
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder config for file (without colors)
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Console writer
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Create cores
	var cores []zapcore.Core

	// Console core (always enabled)
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		consoleWriter,
		zapLevel,
	)
	cores = append(cores, consoleCore)

	// File core (if path provided)
	if logFilePath != "" {
		// Create directory if it doesn't exist
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err == nil {
			// Lumberjack for log rotation
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   logFilePath,
				MaxSize:    10, // MB
				MaxBackups: 3,
				MaxAge:     7, // days
				Compress:   true,
			})

			fileCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(fileEncoderConfig),
				fileWriter,
				zapLevel,
			)
			cores = append(cores, fileCore)
		}
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}

// Desugar returns the underlying zap.Logger
func (l *Logger) Desugar() *zap.Logger {
	return l.SugaredLogger.Desugar()
}
