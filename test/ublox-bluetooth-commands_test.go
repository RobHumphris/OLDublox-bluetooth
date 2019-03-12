package ubloxbluetooth

import (
	"fmt"
	"testing"

	u "github.com/RobHumphris/ublox-bluetooth"
	serial "github.com/RobHumphris/ublox-bluetooth/serial"
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
	for i := 0; i < 1; i++ {
		//exerciseTheDevice("CE1A0B7E9D79r", ub, t, i, true, false)
		exerciseTheDevice("D5926479C652r", ub, t, i, true, false)
		//exerciseTheDevice("C1851F6083F8r", ub, t, i, true, false)
	}
}

func TestAbort(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	// put some events into the sensors logs (assuming connection events are being logged)
	for i := 0; i < 20; i++ {
		fmt.Printf("Connect attempt %d\n", i)
		errr := connectToDevice("D5926479C652r", func(t *testing.T) error {
			return ub.DisconnectFromDevice()
		}, ub, t)
		if errr != nil {
			err = errors.Wrapf(err, "connectToDevice error: %v", errr)
		}
	}

	err = connectToDevice("D5926479C652r", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()
		err = ub.EnableNotifications()
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		info, err := ub.GetInfo()
		if err != nil {
			t.Fatalf("GetInfo error %v\n", err)
		}

		startingIndex := info.CurrentSequenceNumber - info.RecordsCount
		abc := 0
		err = ub.DownloadLogFile(startingIndex, func(b []byte) error {
			abc++
			if abc == 10 {
				fmt.Print("Should Stop\n")
				return ub.AbortEventLogRead()
			}
			return nil
		})
		return err
	}, ub, t)

	if err != nil {
		t.Fatalf("TestAbort error %v\n", err)
	}
}

func TestPagedDownloads(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	// C1851F6083F8r or CE1A0B7E9D79r
	err = connectToDevice("D5926479C652r", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		err := ub.EnableNotifications()
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		info, err := ub.GetInfo()
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current sequence number %d. Records count %d\n", info.CurrentSequenceNumber, info.RecordsCount)
		serial.SetVerbose(true)
		startingIndex := info.CurrentSequenceNumber - info.RecordsCount
		err = ub.DownloadLogFile(startingIndex, func(b []byte) error {
			//fmt.Print(".")
			return nil
		})
		if err != nil {
			t.Errorf("[DownloadLogFile] error: %v\n", err)
		}
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

	err = connectToDevice("CE1A0B7E9D79r", func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		info, err := ub.GetInfo()
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current sequence number %d. Records count %d\n", info.CurrentSequenceNumber, info.RecordsCount)

		err = ub.ClearEventLog()
		if err != nil {
			t.Fatalf("ClearEventLog error %v\n", err)
		}

		info, err = ub.GetInfo()
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current sequence number %d. Records count %d\n", info.CurrentSequenceNumber, info.RecordsCount)
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("TestEventClear error %v\n", err)
	}
}

func TestReboot(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("CE1A0B7E9D79r", func(t *testing.T) error {
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

func setupBluetooth() (*u.UbloxBluetooth, error) {
	ub, err := u.NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		return nil, errors.Wrap(err, "NewUbloxBluetooth error")
	}

	err = ub.ConfigureUblox()
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

func connectToDevice(mac string, tfn TestFunc, ub *u.UbloxBluetooth, t *testing.T) error {
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

func exerciseTheDevice(deviceAddr string, ub *u.UbloxBluetooth, t *testing.T, itteration int, downloadEvents bool, downloadSlotData bool) {
	err := connectToDevice(deviceAddr, func(t *testing.T) error {
		fmt.Printf("[GetInfo] starting\n")
		info, err := ub.GetInfo()
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] replied with: %v\n", info)

		version, err := ub.GetVersion()
		if err != nil {
			t.Fatalf("GetVersion error %v\n", err)
		}
		fmt.Printf("[GetVersion] replied with: %v\n", version)

		config, err := ub.ReadConfig()
		if err != nil {
			t.Fatalf("ReadConfig error %v\n", err)
		}
		fmt.Printf("[ReadConfig] replied with: %v\n", config)

		if downloadEvents {
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
		}

		err = ub.DisconnectFromDevice()
		if err != nil {
			t.Errorf("DisconnectFromDevice Itteration[%d] error %v\n", itteration, err)
		}
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("exerciseTheDevice error %v\n", err)
	}
}
