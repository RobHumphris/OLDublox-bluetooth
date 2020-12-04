package ubloxbluetooth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestDiscovery
func TestDiscoverySingle(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
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
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
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
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
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
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	defer btd.CloseAll()

	btd.SetVerbose(true)

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
	//	cancel()
	fmt.Printf("Ran for %d seconds: %v\n", int32(time.Now().Unix())-timestamp, err)

	ub, err := btd.GetDevice(0)

	for i := 0; i < 2; i++ {
		ub.ConnectToDevice(os.Getenv("DEVICE_MAC"), func(ub *UbloxBluetooth) error {
			err := ub.EnableIndications()
			if err != nil {
				t.Errorf("[EnableIndications] %v\n", err)
			}

			err = ub.EnableNotifications()
			if err != nil {
				t.Errorf("[EnableNotifications] %v\n", err)
			}

			_, err = ub.UnlockDevice(password)
			if err != nil {
				t.Errorf("[UnlockDevice] %v\n", err)
			}

			info, err := ub.GetTime()
			if err != nil {
				t.Errorf("[GetTime] %v\n", err)
			}
			fmt.Printf("[GetTime] replied with: %v\n", info)

			err = ub.DisconnectFromDevice()
			if err != nil {
				t.Errorf("[DisconnectFromDevice] First %v\n", err)
			}

			err = ub.DisconnectFromDevice()
			if err != nil && err.Error() != "ConnectionReply is nil" {
				t.Errorf("[DisconnectFromDevice] Second %v\n", err)
			}
			return nil
		}, func(ub *UbloxBluetooth) error {
			return fmt.Errorf("disconnected")
		})
	}

}
