package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

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

	timestamp := int32(time.Now().Unix())

	alpha := func(dr *DiscoveryReply, timestamp int32) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	err = ub.DiscoveryCommand(timestamp, alpha)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}
}
