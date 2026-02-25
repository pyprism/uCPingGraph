package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pyprism/uCPingGraph/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

var L *zap.Logger

// Init initialises the global zap logger.
// It writes JSON to a rotated file in logsDir and human-readable output to
// stdout.  If SENTRY_DSN is set it also initialises the Sentry SDK.
func Init() {
	logsDir := utils.GetEnv("LOG_DIR", "./logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		log.Fatalf("failed to create logs directory %s: %v", logsDir, err)
	}

	// --- File core (JSON, rotated) ---
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logsDir, "server.log"),
		MaxSize:    50, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), zapcore.InfoLevel)

	// --- Console core (human-readable) ---
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	// Tee both cores together.
	L = zap.New(
		zapcore.NewTee(fileCore, consoleCore),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// --- Sentry (optional) ---
	dsn := utils.GetEnv("SENTRY_DSN", "")
	if dsn != "" {
		env := utils.GetEnv("APP_ENV", "production")
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      env,
			TracesSampleRate: 0.2,
		}); err != nil {
			L.Error("sentry init failed", zap.Error(err))
		} else {
			L.Info("sentry initialised", zap.String("env", env))
		}
	}

	L.Info("logger initialised", zap.String("log_dir", logsDir))
}

// Shutdown flushes the logger and Sentry buffers.
func Shutdown() {
	if L != nil {
		_ = L.Sync()
	}
	sentry.Flush(2 * time.Second)
}

// Get returns the global logger, initialising with a no-op logger if Init was
// not called (e.g. during tests).
func Get() *zap.Logger {
	if L == nil {
		L = zap.NewNop()
	}
	return L
}

// CaptureError logs an error via zap and reports it to Sentry if configured.
func CaptureError(err error, msg string, fields ...zap.Field) {
	l := Get()
	l.Error(msg, append(fields, zap.Error(err))...)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("%s: %w", msg, err))
	}
}
