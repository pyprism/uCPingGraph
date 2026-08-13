package service

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestCleanDBReturnsOnSignal(t *testing.T) {
	setupTestDB(t)

	done := make(chan struct{})
	go func() {
		CleanDB()
		close(done)
	}()

	// Give the goroutine time to register its signal handler and run the
	// initial cleanup pass before we signal it to stop.
	time.Sleep(100 * time.Millisecond)

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("send SIGTERM: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("CleanDB did not return after SIGTERM")
	}
}
