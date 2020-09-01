package ubloxbluetooth

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestDiscovery
func TestDiscoverySingle(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(false)

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	scan := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	timestamp := int32(time.Now().Unix())

	drChan := make(chan *DiscoveryReply, 10)

	go func() {
		for {
			select {
			case dr := <-drChan:
				fmt.Printf("Discovery: %v\n", dr)

			case <-ctx.Done():
				return
			}
		}
	}()

	errChan := make(chan error, 1)
	go ub.discoveryCommandWithContext(ctx, scan, drChan, errChan)
	err = <-errChan
	cancel()
	if err != nil {
		if err == ErrorContextCancelled {
			fmt.Println("function returned ErrorContextCancelled error (which is correct)")
		} else {
			t.Errorf("TestDiscovery(2) error %v\n", err)
		}
	}
	fmt.Printf("Ran for %d seconds\n", int32(time.Now().Unix())-timestamp)
}

func TestDiscoverySingleCancel(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(false)

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	scan := 6 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	timestamp := int32(time.Now().Unix())

	drChan := make(chan *DiscoveryReply, 10)

	go func() {
		for {
			select {
			case dr := <-drChan:
				fmt.Printf("Discovery: %v\n", dr)

			case <-ctx.Done():
				return
			}
		}
	}()

	errChan := make(chan error, 1)
	go ub.discoveryCommandWithContext(ctx, scan, drChan, errChan)
	err = <-errChan
	if err != nil {
		if err == ErrorContextCancelled {
			fmt.Println("function returned ErrorContextCancelled error (which is correct)")
		} else {
			t.Errorf("TestDiscovery(2) error %v\n", err)
		}
	}
	fmt.Printf("2 Ran for %d\n", int32(time.Now().Unix())-timestamp)
}

func TestDiscoveryMulti(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	defer btd.CloseAll()

	btd.SetVerbose(false)

	scan := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	timestamp := int32(time.Now().Unix())

	drChan := make(chan *DiscoveryReply, 10)

	go func() {
		for {
			select {
			case dr := <-drChan:
				fmt.Printf("Discovery: %v\n", dr)

			case <-ctx.Done():
				return
			}
		}
	}()

	err = btd.MultiDiscoverWithContext(ctx, scan, drChan)
	cancel()

	fmt.Printf("Ran for %d seconds\n", int32(time.Now().Unix())-timestamp)

}

func TestDiscoveryMultiCancel(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	defer btd.CloseAll()

	btd.SetVerbose(false)

	scan := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	timestamp := int32(time.Now().Unix())

	drChan := make(chan *DiscoveryReply, 10)

	go func() {
		for {
			select {
			case dr := <-drChan:
				fmt.Printf("Discovery: %v\n", dr)

			case <-ctx.Done():
				return
			}
		}
	}()

	err = btd.MultiDiscoverWithContext(ctx, scan, drChan)
	cancel()

	fmt.Printf("Ran for %d seconds\n", int32(time.Now().Unix())-timestamp)
}
