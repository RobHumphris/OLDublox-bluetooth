package ubloxbluetooth

import (
	"testing"
	"time"
)

func TestATCommand(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
}

// TestDiscovery
func TestDiscovery(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.ConfigureUblox()
	if err != nil {
		t.Fatalf("ConfigureUblox error %v\n", err)
	}

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
