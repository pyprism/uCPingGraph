package logger

import (
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
