package models

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testDB creates an in-memory SQLite DB for the current test and sets the
// global DB variable.  It returns the *gorm.DB for direct queries.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared",
		strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Network{}, &Device{}, &Stat{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	SetDB(db)
	return db
}

// --------------- Network tests ---------------

func TestCreateNetwork(t *testing.T) {
	testDB(t)
	n := Network{}
	id, err := n.CreateNetwork("office")
	if err != nil {
		t.Fatalf("CreateNetwork: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero network ID")
	}
}

func TestCreateNetworkDuplicateNameFails(t *testing.T) {
	testDB(t)
	n1 := Network{}
	if _, err := n1.CreateNetwork("dup"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	n2 := Network{}
	_, err := n2.CreateNetwork("dup")
	if err == nil {
		t.Fatal("expected error for duplicate network name")
	}
}

func TestGetNetworkIdByName(t *testing.T) {
	testDB(t)
	n := Network{}
	id, _ := n.CreateNetwork("lab")

	n2 := Network{}
	got, err := n2.GetNetworkIdByName("lab")
	if err != nil {
		t.Fatalf("GetNetworkIdByName: %v", err)
	}
	if got != id {
		t.Fatalf("expected id %d, got %d", id, got)
	}
}

func TestGetNetworkIdByNameNotFound(t *testing.T) {
	testDB(t)
	n := Network{}
	_, err := n.GetNetworkIdByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing network")
	}
}

func TestGetAllNetwork(t *testing.T) {
	testDB(t)
	for _, name := range []string{"beta", "alpha", "gamma"} {
		n := Network{}
		if _, err := n.CreateNetwork(name); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	n := Network{}
	networks, err := n.GetAllNetwork()
	if err != nil {
		t.Fatalf("GetAllNetwork: %v", err)
	}
	if len(networks) != 3 {
		t.Fatalf("expected 3, got %d", len(networks))
	}
	// Should be ordered ASC
	if networks[0].Name != "alpha" {
		t.Fatalf("expected alpha first, got %s", networks[0].Name)
	}
}

func TestGetAllNetworkName(t *testing.T) {
	testDB(t)
	n := Network{}
	n.CreateNetwork("z-net")
	n2 := Network{}
	n2.CreateNetwork("a-net")

	n3 := Network{}
	names, err := n3.GetAllNetworkName()
	if err != nil {
		t.Fatalf("GetAllNetworkName: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "a-net" {
		t.Fatalf("expected a-net first, got %s", names[0])
	}
}

func TestGetAllNetworkNilDB(t *testing.T) {
	SetDB(nil)
	n := Network{}
	_, err := n.GetAllNetwork()
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

// --------------- Device tests ---------------

func TestCreateDevice(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d := Device{}
	id, token, err := d.CreateDevice(int(nid), "sensor-1")
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero device ID")
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestCreateDeviceDuplicateNameInSameNetwork(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d1 := Device{}
	if _, _, err := d1.CreateDevice(int(nid), "dup-device"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	d2 := Device{}
	_, _, err := d2.CreateDevice(int(nid), "dup-device")
	if err == nil {
		t.Fatal("expected error for duplicate device in same network")
	}
}

func TestCheckDeviceNameIsUnique(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d := Device{}
	if !d.CheckDeviceNameIsUnique(int(nid), "new-device") {
		t.Fatal("expected true for new device name")
	}

	d2 := Device{}
	d2.CreateDevice(int(nid), "existing")

	d3 := Device{}
	if d3.CheckDeviceNameIsUnique(int(nid), "existing") {
		t.Fatal("expected false for existing device name")
	}
}

func TestGetDeviceByToken(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d := Device{}
	expectedID, token, _ := d.CreateDevice(int(nid), "node")

	d2 := Device{}
	id, netID, err := d2.GetDeviceByToken(token)
	if err != nil {
		t.Fatalf("GetDeviceByToken: %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected device id %d, got %d", expectedID, id)
	}
	if netID != int(nid) {
		t.Fatalf("expected network id %d, got %d", nid, netID)
	}
}

func TestGetDeviceByTokenNotFound(t *testing.T) {
	testDB(t)
	d := Device{}
	_, _, err := d.GetDeviceByToken("nonexistent-token")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestGetDevicesByNetwork(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d1 := Device{}
	d1.CreateDevice(int(nid), "z-device")
	d2 := Device{}
	d2.CreateDevice(int(nid), "a-device")

	d3 := Device{}
	devices, err := d3.GetDevicesByNetwork(int(nid))
	if err != nil {
		t.Fatalf("GetDevicesByNetwork: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Name != "a-device" {
		t.Fatalf("expected a-device first, got %s", devices[0].Name)
	}
}

func TestGetDeviceIdByName(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d := Device{}
	expectedID, _, _ := d.CreateDevice(int(nid), "sensor")

	d2 := Device{}
	id, err := d2.GetDeviceIdByName("sensor", nid)
	if err != nil {
		t.Fatalf("GetDeviceIdByName: %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected %d, got %d", expectedID, id)
	}
}

func TestGetDeviceIdByNameNotFound(t *testing.T) {
	testDB(t)
	n := Network{}
	nid, _ := n.CreateNetwork("home")

	d := Device{}
	_, err := d.GetDeviceIdByName("ghost", nid)
	if err == nil {
		t.Fatal("expected error for missing device")
	}
}

func TestDeviceNilDB(t *testing.T) {
	SetDB(nil)
	d := Device{}

	if d.CheckDeviceNameIsUnique(1, "x") {
		t.Fatal("expected false when DB is nil")
	}

	_, _, err := d.CreateDevice(1, "x")
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}

	_, _, err = d.GetDeviceByToken("x")
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}

	_, err = d.GetDevicesByNetwork(1)
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}

	_, err = d.GetDeviceIdByName("x", 1)
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}
