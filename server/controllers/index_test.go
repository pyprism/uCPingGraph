package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHomeRendersDashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// templates/ is loaded relative to the server module root.
	t.Chdir("..")

	router := gin.New()
	router.LoadHTMLGlob("templates/*.html")
	index := new(IndexController)
	router.GET("/", index.Home)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "uCPingGraph Dashboard") {
		t.Fatalf("expected rendered title in body, got %s", rec.Body.String())
	}
}
