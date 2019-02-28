package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	"github.com/RobHumphris/rf-gateway/global"
)

func errorHandler(ub *UbloxBluetooth, t *testing.T) {
	err := ub.ATCommand()
	if err != nil {
		t.Fatalf("ATCommand error %v\n", err)
	}
	fmt.Printf("AT Sent Okay")
}

func retryCall(fn func(*UbloxBluetooth, string) error, ub *UbloxBluetooth, mac string) (err error) {
	for i := 0; i < global.RetryCount; i++ {
		err := fn(ub, mac)
		if err == nil {
			return nil
		}
		global.Debugf("Call failed on device %s, retrying", mac)
		time.Sleep(global.RetryWait)
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
	return ub.ConnectToDevice(deviceAddr, func(cr *ConnectionReply) error {
		fmt.Printf("[ConnectToDevice] replied with: %v\n", cr)
		defer ub.DisconnectFromDevice(cr)

		err := ub.EnableIndications(cr)
		if err != nil {
			return err
		}

		err = ub.EnableNotifications(cr)
		if err != nil {
			return err
		}

		unlocked, err := ub.UnlockDevice(cr, password)
		if err != nil {
			return err
		}
		if !unlocked {
			return err
		}
		fmt.Printf("[UnlockDevice] replied with: %v\n", unlocked)

		version, err := ub.GetVersion(cr)
		if err != nil {
			return err
		}
		fmt.Printf("Software Version: %d Hardware Version: %d", version.SoftwareVersion, version.HardwareVersion)

		info, err := ub.GetInfo(cr)
		if err != nil {
			return err
		}
		fmt.Printf("[GetInfo] replied with: %v\n", info)

		config, err := ub.ReadConfig(cr)
		if err != nil {
			return err
		}
		fmt.Printf("[ReadConfig] replied with: %v\n", config)

		return nil
	}, func(cr *ConnectionReply) error {
		return fmt.Errorf("Disconnected!")
	})
}

func TestSingleAccess(t *testing.T) {
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

	accessDevice(ub, "C1851F6083F8r")
}

func TestMulipleAccesses(t *testing.T) {
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

	for i := 0; i < 100; i++ {
		fmt.Printf("Starting Access test %d\n", i)
		accessDevice(ub, "C1851F6083F8r")
		accessDevice(ub, "CE1A0B7E9D79r")
		accessDevice(ub, "D8CFDFA118ECr")
	}
}
