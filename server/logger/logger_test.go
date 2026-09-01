package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestGetReturnsNonNilLogger(t *testing.T) {
	// Reset global state for test
	mu.Lock()
	L = nil
	mu.Unlock()

	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil")
	}
}

func TestGetReturnsSameInstance(t *testing.T) {
	mu.Lock()
	L = nil
	mu.Unlock()

	l1 := Get()
	l2 := Get()
	if l1 != l2 {
		t.Fatal("Get() returned different instances")
	}
}

func TestGetReturnsInitializedLogger(t *testing.T) {
	custom := zap.NewNop()
	mu.Lock()
	L = custom
	mu.Unlock()

	got := Get()
	if got != custom {
		t.Fatal("Get() did not return the logger set via Init path")
	}
}

func TestCaptureErrorDoesNotPanic(t *testing.T) {
	mu.Lock()
	L = zap.NewNop()
	mu.Unlock()

	// Should not panic even with nil error
	CaptureError(nil, "test message")

	// Should not panic with real error
	CaptureError(errForTest("test error"), "something went wrong")
}

type errForTest string

func (e errForTest) Error() string { return string(e) }

func TestInitCreatesLogDirAndLogger(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "logs")
	t.Setenv("LOG_DIR", logDir)
	t.Setenv("SENTRY_DSN", "")

	Init()
	t.Cleanup(func() {
		mu.Lock()
		L = nil
		mu.Unlock()
	})

	if _, err := os.Stat(logDir); err != nil {
		t.Fatalf("expected log dir to be created: %v", err)
	}

	l := Get()
	l.Info("smoke test entry")
	_ = l.Sync()

	logFile := filepath.Join(logDir, "server.log")
	info, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected log file to contain data")
	}
}

func TestShutdownDoesNotPanicWithoutInit(t *testing.T) {
	mu.Lock()
	L = nil
	mu.Unlock()

	// Should not panic even though Init was never called.
	Shutdown()
}

func TestShutdownFlushesInitializedLogger(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "logs")
	t.Setenv("LOG_DIR", logDir)
	t.Setenv("SENTRY_DSN", "")

	Init()
	t.Cleanup(func() {
		mu.Lock()
		L = nil
		mu.Unlock()
	})

	// Should not panic when flushing a real logger and Sentry's (uninitialised) client.
	Shutdown()
}

func TestInitWithInvalidSentryDSNLogsErrorInsteadOfPanicking(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "logs")
	t.Setenv("LOG_DIR", logDir)
	t.Setenv("SENTRY_DSN", "not-a-valid-dsn")

	Init()
	t.Cleanup(func() {
		mu.Lock()
		L = nil
		mu.Unlock()
	})

	if Get() == nil {
		t.Fatal("expected logger to be initialised even when Sentry DSN is invalid")
	}
}

func TestInitWithValidSentryDSNInitialisesSentry(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "logs")
	t.Setenv("LOG_DIR", logDir)
	t.Setenv("SENTRY_DSN", "https://public@sentry.example.com/1")
	t.Setenv("APP_ENV", "test")

	Init()
	t.Cleanup(func() {
		mu.Lock()
		L = nil
		mu.Unlock()
	})

	if Get() == nil {
		t.Fatal("expected logger to be initialised with a valid Sentry DSN")
	}
}
