package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

var timeout = 6 * time.Second
var password = []byte{'A', 'B', 'C'}

func doConnect(ub *UbloxBluetooth, mac string, t *testing.T) {
	cr, err := ub.ConnectToDevice(mac)
	fmt.Printf("[ConnectToDevice] replied with: %v\n", cr)
	if err != nil {
		t.Errorf("TestConnect error %v\n", err)
	}
	defer ub.DisconnectFromDevice(cr)

	if cr.BluetoothAddress != mac {
		t.Errorf("ConnectToDevice - addresses do not match")
	}
	if cr.Type != 0 {
		t.Errorf("ConnectToDevice - type is unknown should be zero")
	}
}

func TestMultipleConnects(t *testing.T) {
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

	for i := 0; i < 100; i++ {
		fmt.Printf("Starting connect test %d\n", i)
		doConnect(ub, "C1851F6083F8r", t)
		doConnect(ub, "CE1A0B7E9D79r", t)
		doConnect(ub, "D8CFDFA118ECr", t)
	}
}
