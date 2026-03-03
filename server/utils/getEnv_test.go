package utils

import (
	"os"
	"testing"
)

func TestGetEnvReturnsDefault(t *testing.T) {
	key := "TEST_GETENV_MISSING_KEY_12345"
	os.Unsetenv(key)

	got := GetEnv(key, "fallback")
	if got != "fallback" {
		t.Fatalf("expected 'fallback', got %q", got)
	}
}

func TestGetEnvReturnsSetValue(t *testing.T) {
	key := "TEST_GETENV_SET_KEY_12345"
	t.Setenv(key, "custom_value")

	got := GetEnv(key, "fallback")
	if got != "custom_value" {
		t.Fatalf("expected 'custom_value', got %q", got)
	}
}

func TestGetEnvEmptyStringReturnsDefault(t *testing.T) {
	key := "TEST_GETENV_EMPTY_KEY_12345"
	t.Setenv(key, "")

	got := GetEnv(key, "default")
	if got != "default" {
		t.Fatalf("expected 'default' for empty env, got %q", got)
	}
}
