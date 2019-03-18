package ubloxbluetooth

import (
	"fmt"
	"testing"

	u "github.com/RobHumphris/ublox-bluetooth"
	serial "github.com/RobHumphris/ublox-bluetooth/serial"
)

// TestDiscovery
func TestDiscovery(t *testing.T) {
	serial.SetVerbose(true)
	ub, err := u.NewUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	alpha := func(dr *u.DiscoveryReply) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	err = ub.DiscoveryCommand(alpha)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}
}
