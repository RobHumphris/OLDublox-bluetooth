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

	serial.SetVerbose(true)

	err = connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
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

		/*if downloadEvents {
			startingIndex := info.CurrentSequenceNumber - info.RecordsCount
			fmt.Printf("[DownloadLogFile] starting run: %d\n", itteration)
			err = ub.DownloadLogFile(startingIndex, func(b []byte) error {
				//fmt.Print(".")
				return nil
			})
			if err != nil {
				t.Fatalf("DownloadLogFile error %v\n", err)
			}
			fmt.Printf("[DownloadLogFile] complete\n")
		}

		if downloadSlotData {
			slotCount, err := ub.ReadSlotCount()
			if err != nil {
				t.Errorf("ReadSlotCount error %v\n", err)
			} else {
				fmt.Printf("[ReadSlotCount] replied with: %v\n", slotCount)
				slotInfo, err := ub.ReadSlotInfo(0)
				if err != nil {
					t.Errorf("ReadSlotInfo error %v\n", err)
				} else {
					fmt.Printf("[ReadSlotInfo] replied with: %v\n", slotInfo)
					slotData, err := ub.ReadSlotData(0, 0, slotInfo.Bytes)
					if err != nil {
						t.Errorf("ReadSlotData error %v\n", err)
					}
					fmt.Printf("[ReadSlotData] replied with: %v\n", slotData)
				}
			}
		}*/

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

func TestEventClear(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		t.Fatalf("TODO - rework")

		time, err := ub.GetTime()
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current timestamp %d\n", time)

		/*err = ub.ClearEventLog()
		if err != nil {
			t.Fatalf("ClearEventLog error %v\n", err)
		}*/
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("TestEventClear error %v\n", err)
	}
}

func TestSlotDataClear(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		startingCount, err := ub.ReadSlotCount()
		if err != nil {
			t.Fatalf("ReadSlotCount error %v\n", err)
		}

		err = ub.EraseSlotData()
		if err != nil {
			t.Fatalf("EraseSlotData error %v\n", err)
		}

		newCount, err := ub.ReadSlotCount()
		if err != nil {
			t.Fatalf("ReadSlotCount error %v\n", err)
		}

		fmt.Printf("Original count %d, count after erase %d\n", startingCount.Count, newCount.Count)

		return err
	}, ub, t)

	if err != nil {
		t.Errorf("TestSlotDataClear error %v\n", err)
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
