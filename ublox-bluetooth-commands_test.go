package ubloxbluetooth

import (
	"fmt"
	"testing"

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
		//exerciseTheDevice("D5926479C652r", ub, t, i, true, false)
		exerciseTheDevice("C1851F6083F8r", ub, t, i, true, false)
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
		errr := connectToDevice("D5926479C652r", func(cr *ConnectionReply, t *testing.T) error {
			return ub.DisconnectFromDevice(cr)
		}, ub, t)
		if errr != nil {
			err = errors.Wrapf(err, "connectToDevice error: %v", errr)
		}
	}

	err = connectToDevice("D5926479C652r", func(cr *ConnectionReply, t *testing.T) error {
		defer ub.DisconnectFromDevice(cr)
		err = ub.EnableNotifications(cr)
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		info, err := ub.GetInfo(cr)
		if err != nil {
			t.Fatalf("GetInfo error %v\n", err)
		}

		startingIndex := info.CurrentSequenceNumber - info.RecordsCount
		abc := 0
		err = ub.DownloadLogFile(cr, startingIndex, func(b []byte) error {
			abc++
			if abc == 10 {
				fmt.Print("Should Stop\n")
				return ub.AbortEventLogRead(cr)
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
	err = connectToDevice("D5926479C652r", func(cr *ConnectionReply, t *testing.T) error {
		defer ub.DisconnectFromDevice(cr)

		err := ub.EnableNotifications(cr)
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		info, err := ub.GetInfo(cr)
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current sequence number %d. Records count %d\n", info.CurrentSequenceNumber, info.RecordsCount)
		serial.SetVerbose(true)
		startingIndex := info.CurrentSequenceNumber - info.RecordsCount
		err = ub.DownloadLogFile(cr, startingIndex, func(b []byte) error {
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

	err = connectToDevice("CE1A0B7E9D79r", func(cr *ConnectionReply, t *testing.T) error {
		defer ub.DisconnectFromDevice(cr)

		info, err := ub.GetInfo(cr)
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] Current sequence number %d. Records count %d\n", info.CurrentSequenceNumber, info.RecordsCount)

		err = ub.ClearEventLog(cr)
		if err != nil {
			t.Fatalf("ClearEventLog error %v\n", err)
		}

		info, err = ub.GetInfo(cr)
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

	err = connectToDevice("CE1A0B7E9D79r", func(cr *ConnectionReply, t *testing.T) error {
		defer ub.DisconnectFromDevice(cr)
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
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
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

type TestFunc func(*ConnectionReply, *testing.T) error

func connectToDevice(mac string, tfn TestFunc, ub *UbloxBluetooth, t *testing.T) error {
	return ub.ConnectToDevice(mac, func(cr *ConnectionReply) error {
		fmt.Printf("[ConnectionReply] %v\n", cr)

		err := ub.EnableNotifications(cr)
		if err != nil {
			t.Fatalf("EnableNotifications error %v\n", err)
		}

		err = ub.EnableIndications(cr)
		if err != nil {
			t.Fatalf("EnableIndications error %v\n", err)
		}

		unlocked, err := ub.UnlockDevice(cr, password)
		if err != nil {
			t.Fatalf("UnlockDevice error %v\n", err)
		}
		if !unlocked {
			t.Fatalf("UnlockDevice error - failed to unlock")
		}

		return tfn(cr, t)
	}, func(cr *ConnectionReply) error {
		fmt.Println("Disconnected!")
		return nil
	})
}

func exerciseTheDevice(deviceAddr string, ub *UbloxBluetooth, t *testing.T, itteration int, downloadEvents bool, downloadSlotData bool) {
	err := connectToDevice(deviceAddr, func(cr *ConnectionReply, t *testing.T) error {
		fmt.Printf("[GetInfo] starting\n")
		info, err := ub.GetInfo(cr)
		if err != nil {
			t.Errorf("GetInfo error %v\n", err)
		}
		fmt.Printf("[GetInfo] replied with: %v\n", info)

		version, err := ub.GetVersion(cr)
		if err != nil {
			t.Fatalf("GetVersion error %v\n", err)
		}
		fmt.Printf("[GetVersion] replied with: %v\n", version)

		config, err := ub.ReadConfig(cr)
		if err != nil {
			t.Fatalf("ReadConfig error %v\n", err)
		}
		fmt.Printf("[ReadConfig] replied with: %v\n", config)

		if downloadEvents {
			startingIndex := info.CurrentSequenceNumber - info.RecordsCount
			fmt.Printf("[DownloadLogFile] starting run: %d\n", itteration)
			err = ub.DownloadLogFile(cr, startingIndex, func(b []byte) error {
				//fmt.Print(".")
				return nil
			})
			if err != nil {
				t.Fatalf("DownloadLogFile error %v\n", err)
			}
			fmt.Printf("[DownloadLogFile] complete\n")
		}

		if downloadSlotData {
			slotCount, err := ub.ReadSlotCount(cr)
			if err != nil {
				t.Errorf("ReadSlotCount error %v\n", err)
			} else {
				fmt.Printf("[ReadSlotCount] replied with: %v\n", slotCount)
				slotInfo, err := ub.ReadSlotInfo(cr, 0)
				if err != nil {
					t.Errorf("ReadSlotInfo error %v\n", err)
				} else {
					fmt.Printf("[ReadSlotInfo] replied with: %v\n", slotInfo)
					slotData, err := ub.ReadSlotData(cr, 0, 0, slotInfo.Bytes)
					if err != nil {
						t.Errorf("ReadSlotData error %v\n", err)
					}
					fmt.Printf("[ReadSlotData] replied with: %v\n", slotData)
				}
			}
		}

		err = ub.DisconnectFromDevice(cr)
		if err != nil {
			t.Errorf("DisconnectFromDevice Itteration[%d] error %v\n", itteration, err)
		}
		return err
	}, ub, t)

	if err != nil {
		t.Errorf("exerciseTheDevice error %v\n", err)
	}
}
