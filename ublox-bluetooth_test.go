package ubloxbluetooth

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestATCommand(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(true)

	err = ub.EnterExtendedDataMode()
	if err != nil {
		t.Errorf("EnterExtendedDataMode error %v\n", err)
	}

	err = ub.EchoOff()
	if err != nil {
		t.Errorf("EchoOff error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	settings, err := ub.GetRS232Settings()
	if err != nil {
		t.Errorf("GetRS232Settings %v\n", err)
	}
	fmt.Printf("RS232 Settings: %v\n", settings)

	err = ub.RebootUblox()
	if err != nil {
		t.Errorf("RebootUblox error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
}

func TestMultiPortInitialisation(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	for idx, ub := range btd.Bt {
		sp := ub.GetSerialPort()
		sn := ub.GetSerialNumber()
		di := ub.GetDeviceIndex()

		if int(di) != idx {
			t.Errorf("Device Index mismatch: %v != %v\n", idx, di)
		}

		ub2, err := btd.GetDevice(int(di))
		if err != nil {
			t.Errorf("GetDevice() error: %v\n", err)
		}

		if ub2 != ub {
			t.Errorf("Device handles don't match: %v != %v\n", ub, ub2)
		}

		fmt.Printf("BT Dongle %v, %v, %v\n", di, sp, sn)
	}

	_, err = btd.GetDevice(-1)
	if err == nil || err.Error() != ErrBadDeviceIndex.Error() {
		t.Errorf("GetDevice() did not return the correct error: %v != %v\n", ErrBadDeviceIndex, err)
	}

	_, err = btd.GetDevice(btd.DeviceCount())
	if err == nil || err.Error() != ErrBadDeviceIndex.Error() {
		t.Errorf("GetDevice() did not return the correct error: %v != %v\n", ErrBadDeviceIndex, err)
	}
}

func TestDualATControl(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	if btd.DeviceCount() == 2 {
		ub0, err := btd.GetDevice(0)
		ub1, err := btd.GetDevice(1)

		defer func() {
			ub0.Close()
			ub1.Close()
		}()

		btd.SetVerbose(true)

		err = ub0.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}

		err = ub1.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}

		err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
			defer ub0.DisconnectFromDevice()

			currentTime := int32(time.Now().Unix())
			deviceTime, err := ub0.GetTime()
			if err != nil {
				return errors.Wrap(err, "GetTime error")
			}

			fmt.Printf("Current time: %d, device time: %d\n", currentTime, deviceTime)

			timeAdjust, err := ub0.SetTime(currentTime)
			if err != nil {
				return errors.Wrap(err, "SetTime error")
			}

			fmt.Printf("TimeAdjustReply CurrentTime: %d UpdatedTime: %d\n", timeAdjust.CurrentTime, timeAdjust.UpdatedTime)
			return nil
		}, ub0, t)
		if err != nil {
			t.Errorf("TestReboot error %v\n", err)
		}

		err = ub0.RebootUblox()
		if err != nil {
			t.Errorf("RebootUblox error %v\n", err)
		}

		err = ub1.RebootUblox()
		if err != nil {
			t.Errorf("RebootUblox error %v\n", err)
		}

		err = ub0.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}

		err = ub1.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}

		err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
			defer ub1.DisconnectFromDevice()
			rssi, er := ub1.GetRSSI()
			if er != nil {
				return er
			}

			fmt.Printf("RSSI Channel: %d dbm: %d\n", rssi.Channel, rssi.Dbm)
			return nil
		}, ub1, t)
		if err != nil {
			t.Errorf("TestReboot error %v\n", err)
		}

		//btd.CloseAll()
	} else {
		t.Errorf("This test needs two EVKs/EH750s plugged in to work")
	}
}

func TestGetSerialNumber(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(true)

	sn, err := ub.getSerialNumber()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
	fmt.Printf("Serial No: %v\n", sn)

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	sn2 := ub.GetSerialNumber()

	sn = strings.Trim(sn, "\"")
	if sn != sn2 {
		t.Errorf("Serial number mismatch: %v != %v\n", sn, sn2)
	}
}
