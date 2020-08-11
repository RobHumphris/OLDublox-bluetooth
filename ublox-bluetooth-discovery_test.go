package ubloxbluetooth

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestDiscovery
func TestDiscovery(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(true)

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	timestamp := int32(time.Now().Unix())

	alpha := func(dr *DiscoveryReply, timestamp int32) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	scan := 6 * time.Second
	/*
		err = ub.DiscoveryCommand(timestamp, scan, alpha)
		if err != nil {
			t.Errorf("TestDiscovery(1) error %v\n", err)
		}
		fmt.Printf("1 Ran for %d\n", int32(time.Now().Unix())-timestamp)
	*/
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	timestamp = int32(time.Now().Unix())
	err = ub.DiscoveryCommandWithContext(ctx, timestamp, scan, alpha)
	if err != nil {
		if err == ErrorContextCancelled {
			fmt.Println("function returned ErrorContextCancelled error (which is correct)")
		} else {
			t.Errorf("TestDiscovery(2) error %v\n", err)
		}

	}
	fmt.Printf("2 Ran for %d\n", int32(time.Now().Unix())-timestamp)
}
