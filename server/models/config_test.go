package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConnectDbWithPathCreatesSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "test.db")

	if err := ConnectDbWithPath(dbPath); err != nil {
		t.Fatalf("ConnectDbWithPath: %v", err)
	}
	t.Cleanup(func() { SetDB(nil) })

	if DB == nil {
		t.Fatal("expected DB to be set")
	}

	if !DB.Migrator().HasTable(&Network{}) {
		t.Fatal("expected networks table to exist after connect")
	}
	if !DB.Migrator().HasTable(&Device{}) {
		t.Fatal("expected devices table to exist after connect")
	}
	if !DB.Migrator().HasTable(&Stat{}) {
		t.Fatal("expected stats table to exist after connect")
	}

	n := Network{}
	if _, err := n.CreateNetwork("test"); err != nil {
		t.Fatalf("sanity write through connected DB: %v", err)
	}
}

func TestConnectDbWithPathFailsWhenDirCannotBeCreated(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}

	// blocker is a regular file, so MkdirAll for a path underneath it fails.
	dbPath := filepath.Join(blocker, "sub", "test.db")
	if err := ConnectDbWithPath(dbPath); err == nil {
		t.Fatal("expected error when db directory cannot be created")
	}
}

func TestConnectDbUsesDbPathEnv(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "env-test.db")
	t.Setenv("DB_PATH", dbPath)

	if err := ConnectDb(); err != nil {
		t.Fatalf("ConnectDb: %v", err)
	}
	t.Cleanup(func() { SetDB(nil) })

	if DB == nil {
		t.Fatal("expected DB to be set")
	}
}
