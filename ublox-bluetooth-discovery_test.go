package ubloxbluetooth

import (
	"fmt"
	"testing"

	serial "github.com/8power/ublox-bluetooth/serial"
)

// TestDiscovery
func TestDiscovery(t *testing.T) {
	serial.SetVerbose(true)
	ub, err := NewUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

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
}
