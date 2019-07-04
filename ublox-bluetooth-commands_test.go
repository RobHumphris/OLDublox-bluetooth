package ubloxbluetooth

import (
	"fmt"
	"testing"

	serial "github.com/8power/ublox-bluetooth/serial"
	"github.com/pkg/errors"
)

const serviceUUIDLength = 42
const serviceUUIDHeaderLength = 10
const serviceUUID = "23E1B7EA5F782315A7BEADDE10138888"

// TestUbloxBluetoothCommands treads through the list of implemented commands
func TestUbloxBluetoothCommands(t *testing.T) {
	serial.SetVerbose(false)
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	//serial.SetVerbose(true)

	err = connectToDevice("CE1A0B7E9D79r", func(t *testing.T) error {
		fmt.Printf("[GetTime] starting\n")
		time, err := ub.GetTime()
		if err != nil {
			t.Errorf("GetTime error %v\n", err)
		}
		fmt.Printf("[GetTime] Current timestamp %d\n", time)

		version, err := ub.GetVersion()
		if err != nil {
			t.Fatalf("GetVersion error %v\n", err)
		}
		fmt.Printf("[GetVersion] replied with: %v\n", version)

		if version.SoftwareVersion != "2.0" {
			t.Fatalf("Cannot continue with version %s, needs to be 2.0\n", version.SoftwareVersion)
		}

		config, err := ub.ReadConfig()
		if err != nil {
			t.Fatalf("ReadConfig error %v\n", err)
		}
		fmt.Printf("[ReadConfig] replied with: %v\n", config)

		echo, err := ub.EchoCommand("012345678901234567890123456789012345678901234567890123456789")
		if err != nil {
			t.Fatalf("EchoCommand error %v\n", err)
		}
		fmt.Printf("[EchoCommand] replied with: %v\n", echo)

		info, err := ub.ReadRecorderInfo()
		if err != nil {
			t.Fatalf("ReadRecorderInfo error %v\n", err)
		}
		fmt.Printf("[ReadRecorderInfo] replied with: %v\n", info)

		lastSequenceRead := 0
		dataSequences := []int{}
		err = ub.ReadRecorder(info.SequenceNo-info.Count, func(e *VehEvent) error {

			lastSequenceRead = e.Sequence
			if e.DataFlag {
				dataSequences = append(dataSequences, e.Sequence)
			}
			return nil
		})
		if err != nil {
			t.Errorf("ReadRecorder error %v\n", err)
		}
		fmt.Printf("[ReadRecorder] Final Sequence %d events\n", lastSequenceRead)
		fmt.Printf("[ReadRecorder] has %d data sequences to download\n", len(dataSequences))

		err = ub.DisconnectFromDevice()
		if err != nil {
			t.Errorf("DisconnectFromDevice error %v\n", err)
		}
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("exerciseTheDevice error %v\n", err)
	}

}

func TestPagedDownloads(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		err := ub.EnableNotifications()
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		time, err := ub.GetTime()
		if err != nil {
			t.Errorf("GetTime error %v\n", err)
		}
		fmt.Printf("[GetTime] Current timestamp %d\n", time)
		serial.SetVerbose(true)
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("TestPagedDownloads error %v\n", err)
	}
}

func TestReboot(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()
		ub.PeerList()
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("TestReboot error %v\n", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Errorf("RebootUblox error %v\n", err)
	}
	fmt.Printf("Rebooted")
}

func setupBluetooth() (*UbloxBluetooth, error) {
	ub, err := NewUbloxBluetooth(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "NewUbloxBluetooth error")
	}

	/*err = ub.ConfigureUblox()
	if err != nil {
		return nil, errors.Wrap(err, "ConfigureUblox error")
	}*/

	err = ub.RebootUblox()
	if err != nil {
		return nil, errors.Wrap(err, "RebootUblox error")
	}

	err = ub.ATCommand()
	if err != nil {
		return nil, errors.Wrap(err, "AT error")
	}

	err = ub.EchoOff()
	if err != nil {
		return nil, errors.Wrap(err, "EchoOff error")
	}

	err = ub.ATCommand()
	if err != nil {
		return nil, errors.Wrap(err, "AT error")
	}

	return ub, nil
}

type TestFunc func(*testing.T) error

func connectToDevice(mac string, tfn TestFunc, ub *UbloxBluetooth, t *testing.T) error {
	return ub.ConnectToDevice(mac, func() error {
		err := ub.EnableNotifications()
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		err = ub.EnableIndications()
		if err != nil {
			t.Fatalf("EnableIndications error %v\n", err)
		}

		unlocked, err := ub.UnlockDevice(password)
		if err != nil {
			t.Fatalf("UnlockDevice error %v\n", err)
		}
		if !unlocked {
			t.Fatalf("UnlockDevice error - failed to unlock")
		}

		return tfn(t)
	}, func() error {
		fmt.Println("Disconnected!")
		return nil
	})
}
