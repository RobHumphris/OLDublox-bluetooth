package ubloxbluetooth

import (
	"fmt"
	"testing"

	serial "github.com/8power/ublox-bluetooth/serial"
)

func TestRSSICommand(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	serial.SetVerbose(true)
	doRSSITest("CE1A0B7E9D79r", ub, t)

}

func doRSSITest(address string, ub *UbloxBluetooth, t *testing.T) {
	err := connectToDevice(address, func(t *testing.T) error {
		rssi, er := ub.GetRSSI()
		if er != nil {
			return er
		}

		fmt.Printf("RSSI Channel: %d dbm: %d\n", rssi.Channel, rssi.Dbm)
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("Connect to device error %v\n", err)
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
