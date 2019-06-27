package ubloxbluetooth

import (
	"fmt"
	"testing"

	serial "github.com/RobHumphris/ublox-bluetooth/serial"
)

func TestRSSICommand(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	serial.SetVerbose(true)
	doRSSITest("CE1A0B7E9D79", ub, t)
	doRSSITest("D5926479C652r", ub, t)
	doRSSITest("C1851F6083F8r", ub, t)
}

func doRSSITest(address string, ub *UbloxBluetooth, t *testing.T) {
	rssi, err := ub.GetDeviceRSSI(address)
	if err != nil {
		t.Errorf("GetDeviceRSSI %s error %v\n", address, err)
	} else {
		fmt.Printf("%s RSSI: %s", address, rssi)
	}
	rssi, err = ub.GetDeviceRSSI(address)
	if err != nil {
		t.Errorf("GetDeviceRSSI %s error %v\n", address, err)
	} else {
		fmt.Printf("%s RSSI: %s", address, rssi)
	}
}

func TestPeerList(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()
	err = ub.PeerList()
	if err != nil {
		t.Fatalf("PeerList error %v\n", err)
	}
}
