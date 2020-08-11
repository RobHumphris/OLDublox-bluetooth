package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/pkg/errors"
)

const serviceUUIDLength = 42
const serviceUUIDHeaderLength = 10
const serviceUUID = "23E1B7EA5F782315A7BEADDE10138888"

// TestUbloxBluetoothCommands treads through the list of implemented commands
func TestUbloxBluetoothCommands(t *testing.T) {
	defer leaktest.Check(t)()
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()
	ub.serialPort.SetVerbose(true)

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		version, err := ub.GetVersion()
		if err != nil {
			t.Fatalf("GetVersion error %v\n", err)
		}
		fmt.Printf("[GetVersion] replied with: %v\n", version)

		serialNumber, err := ub.ReadSerialNumber()
		if err != nil {
			t.Fatalf("ReadSerialNumber error %v\n", err)
		}
		fmt.Printf("[ReadSerialNumber] replied with: %s\n", serialNumber)

		fmt.Printf("[GetTime] starting\n")
		time, err := ub.GetTime()
		if err != nil {
			t.Errorf("GetTime error %v\n", err)
		}
		fmt.Printf("[GetTime] Current timestamp %d\n", time)

		/*if version.SoftwareVersion != "3.0" {
			t.Fatalf("Cannot continue with version %s, needs to be 2.1\n", version.SoftwareVersion)
		}*/

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
		fmt.Printf("[ReadRecorderInfo] SequenceNo: %d. Count: %d. SlotUsage: %d. PoolUsage: %d.\n", info.SequenceNo, info.Count, info.SlotUsage, info.PoolUsage)

		var lastSequenceRead uint32
		dataSequences := []uint32{}

		err = ub.ReadRecorder(0, func(e *VehEvent) error {
			fmt.Printf("Sequence: %d\n", e.Sequence)
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

		for _, s := range dataSequences {
			meta, err := ub.QueryRecorderMetaDataCommand(s)
			if err != nil {
				t.Errorf("QueryRecorderMetaDataCommand error %v", err)
			} else {
				fmt.Printf("Metadata - Valid: %t\tLength: %d\tCRC: %X", meta.Valid, meta.Length, meta.Crc)
				if meta.Valid {

				}
			}
		}

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

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
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
		ub.serialPort.SetVerbose(true)
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("TestPagedDownloads error %v\n", err)
	}
}

func TestAttemptToConnectToMissing(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("EEEEEEEEEEEEr", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()
		return nil
	}, ub, t)
	if err.Error() != "Timeout" {
		t.Errorf("TestReboot error %v\n", err)
	}

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		defer ub.DisconnectFromDevice()
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("TestReboot error %v\n", err)
	}
}

func TestRebootUblox(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		defer ub.DisconnectFromDevice()
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
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "NewUbloxBluetooth error")
	}

	ub, err := btd.GetDevice(0)

	err = ub.ConfigureUblox(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "ConfigureUblox error")
	}

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
