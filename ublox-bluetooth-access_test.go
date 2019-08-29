package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	serial "github.com/8power/ublox-bluetooth/serial"
	"github.com/google/martian/log"
)

func errorHandler(ub *UbloxBluetooth, t *testing.T) {
	err := ub.ATCommand()
	if err != nil {
		t.Fatalf("ATCommand error %v\n", err)
	}
	fmt.Printf("AT Sent Okay")
}

func retryCall(fn func(*UbloxBluetooth, string) error, ub *UbloxBluetooth, mac string) (err error) {
	for i := 0; i < 3; i++ {
		err := fn(ub, mac)
		if err == nil {
			return nil
		}
		log.Debugf("Call failed on device %s, retrying", mac)
		time.Sleep(500 * time.Millisecond)
		e := ub.ATCommand()
		if e != nil {
			// we cannot continue
			return err
		}
	}
	return err
}

func accessDevice(ub *UbloxBluetooth, mac string) error {
	err := retryCall(accessDeviceFn, ub, mac)
	return err
}

func accessDeviceFn(ub *UbloxBluetooth, deviceAddr string) error {
	return ub.ConnectToDevice(deviceAddr, func() error {
		defer ub.DisconnectFromDevice()

		err := ub.EnableIndications()
		if err != nil {
			return err
		}

		err = ub.EnableNotifications()
		if err != nil {
			return err
		}

		unlocked, err := ub.UnlockDevice(password)
		if err != nil {
			return err
		}
		if !unlocked {
			return err
		}
		fmt.Printf("[UnlockDevice] replied with: %v\n", unlocked)

		version, err := ub.GetVersion()
		if err != nil {
			return err
		}
		fmt.Printf("[GetVersion] Software Version: %s Hardware Version: %s\n", version.SoftwareVersion, version.HardwareVersion)

		info, err := ub.GetTime()
		if err != nil {
			return err
		}
		fmt.Printf("[GetTime] replied with: %v\n", info)

		config, err := ub.ReadConfig()
		if err != nil {
			return err
		}
		fmt.Printf("[ReadConfig] replied with: %v\n", config)

		return nil
	}, func() error {
		return fmt.Errorf("disconnected")
	})
}

func TestSingleAccess(t *testing.T) {
	serial.SetVerbose(true)
	ub, err := NewUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.EnterExtendedDataMode()
	if err != nil {
		t.Fatalf("EnterDataMode error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	accessDevice(ub, "EE9EF8BA058Br")
}

func TestMulipleAccesses(t *testing.T) {
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

	for i := 0; i < 1; i++ {
		//fmt.Printf("Starting Access test %d\n", i)
		//t.Fatalf("NEED MORE v2.0 sensors")
		//accessDevice(ub, "C1851F6083F8r")
		accessDevice(ub, "CE1A0B7E9D79r")
		//accessDevice(ub, "D8CFDFA118ECr")
	}
}
