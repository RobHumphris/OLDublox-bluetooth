package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func TestDoubleDisconnect(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()
	ub.serialPort.SetVerbose(true)

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

func accessDeviceFn(ub *UbloxBluetooth, deviceAddr string) error {
	return ub.ConnectToDevice(deviceAddr, func(ub *UbloxBluetooth) error {
		defer ub.DisconnectFromDevice()
		ub.serialPort.SetVerbose(false)

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
	}, func(ub *UbloxBluetooth) error {
		return fmt.Errorf("disconnected")
	})
}

func TestSingleAccess(t *testing.T) {
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

	accessDevice(ub, os.Getenv("DEVICE_MAC"))

}

func TestMulipleAccesses(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()
	ub.serialPort.SetVerbose(true)

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT errors")
	}

	for i := 0; i < 10; i++ {
		//fmt.Printf("Starting Access test %d\n", i)
		//t.Fatalf("NEED MORE v2.0 sensors")
		//accessDevice(ub, "C1851F6083F8r")
		accessDevice(ub, os.Getenv("DEVICE_MAC"))
		//accessDevice(ub, "D8CFDFA118ECr")
	}
}
