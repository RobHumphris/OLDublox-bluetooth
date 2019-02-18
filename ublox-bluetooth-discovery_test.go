package ubloxbluetooth

import (
	"fmt"
	"testing"
)

// TestDiscovery
func TestDiscovery(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB1", timeout)
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

	alpha := func(dr *DiscoveryReply) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	err = ub.DiscoveryCommand(alpha)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}

	err = ub.DiscoveryCommand(alpha)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}

}
