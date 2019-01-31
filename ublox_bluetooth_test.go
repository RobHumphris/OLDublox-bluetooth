package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

var timeout = 6 * time.Second
var password = []byte{'A', 'B', 'C'}

func TestFloom(tt *testing.T) {
}

// TestDiscovery
func TestDiscovery(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
	time.Sleep(1 * time.Second)

	discovered, err := ub.DiscoveryCommand()
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}
	if len(discovered) < 1 {
		t.Errorf("No discovered devices found\n")
	}
	time.Sleep(5 * time.Second)

	discovered, err = ub.DiscoveryCommand()
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}
	if len(discovered) < 1 {
		t.Errorf("No discovered devices found\n")
	}
}

// TestUbloxBluetoothCommands treads through the list of implemented commands
func TestUbloxBluetoothCommands(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
	time.Sleep(1 * time.Second)

	deviceAddr := "C1851F6083F8r"
	//deviceAddr := "CE1A0B7E9D79r"
	cr, err := ub.ConnectToDevice(deviceAddr)

	fmt.Printf("[ConnectToDevice] replied with: %v\n", cr)
	if err != nil {
		t.Errorf("TestConnect error %v\n", err)
	}
	if cr.BluetoothAddress != deviceAddr {
		t.Errorf("ConnectToDevice - addresses do not match")
	}
	if cr.Type != 0 {
		t.Errorf("ConnectToDevice - type is unknown should be zero")
	}
	time.Sleep(1 * time.Second)

	err = ub.EnableIndications(cr)
	if err != nil {
		t.Errorf("EnableIndications error %v\n", err)
	}

	unlocked, err := ub.UnlockDevice(cr, password)
	if err != nil {
		t.Errorf("UnlockDevice error %v\n", err)
	}
	if !unlocked {
		t.Errorf("UnlockDevice error - failed to unlock")
	}
	fmt.Printf("[UnlockDevice] replied with: %v\n", unlocked)
	time.Sleep(1 * time.Second)

	version, err := ub.GetVersion(cr)
	if err != nil {
		t.Errorf("GetVersion error %v\n", err)
	}
	fmt.Printf("[GetVersion] replied with: %v\n", version)

	info, err := ub.GetInfo(cr)
	if err != nil {
		t.Errorf("GetInfo error %v\n", err)
	}
	fmt.Printf("[GetInfo] replied with: %v\n", info)

	config, err := ub.ReadConfig(cr)
	if err != nil {
		t.Errorf("ReadConfig error %v\n", err)
	}
	fmt.Printf("[ReadConfig] replied with: %v\n", config)

	err = ub.DisconnectFromDevice(cr)
	if err != nil {
		t.Errorf("DisconnectFromDevice error %v\n", err)
	}
}
