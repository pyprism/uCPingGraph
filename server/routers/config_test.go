package routers

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestInitServesAndShutsDownGracefullyOnSIGTERM(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.SetDB(routersTestDB(t))
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	port := freePort(t)
	t.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))

	done := make(chan struct{})
	go func() {
		Init()
		close(done)
	}()

	addr := fmt.Sprintf("http://127.0.0.1:%d/healthz", port)
	var resp *http.Response
	var err error
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = http.Get(addr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("server did not become reachable: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from healthz, got %d", resp.StatusCode)
	}

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("send SIGTERM: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Init did not shut down within timeout after SIGTERM")
	}

	if _, err := http.Get(addr); err == nil {
		t.Fatal("expected server to stop accepting connections after shutdown")
	}
}
