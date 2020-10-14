package ubloxbluetooth

import (
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestSettingDTR(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(true)
	err = ub.SetDTRBehavior()
	if err != nil {
		t.Fatalf("SetDTRBehavior error %v", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v", err)
	}

	err = ub.ResetUblox()
	if err != nil {
		t.Fatalf("SetDTRBehavior error %v", err)
	}
}

func TestSettingDTRAction(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	ub.serialPort.SetVerbose(true)
	err = ub.MultipleATCommands()
	if err != nil {
		t.Fatalf("MultipleATCommands error %v", err)
	}

	err = ub.SetDTRBehavior()
	if err != nil {
		t.Fatalf("SetDTRBehavior error %v", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v", err)
	}

	err = ub.ResetUblox()
	if err != nil {
		t.Fatalf("SetDTRBehavior error %v", err)
	}
}

func TestDongleReboot(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)
	ub.serialPort.SetVerbose(true)
	defer ub.Close()
}

func TestRebootRecorder(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		ub.serialPort.SetVerbose(true)
		e := ub.RebootRecorder()
		if e != nil {
			return errors.Wrap(e, "RebootRecorder error")
		}

		e = ub.DisconnectFromDevice()
		if e != nil {
			return errors.Wrap(e, "DisconnectFromDevice error")
		}
		return nil
	}, ub, t)

	if err != nil {
		t.Errorf("TestReboot error %v\n", err)
	}
}

func TestErase(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		ub.serialPort.SetVerbose(true)
		e := ub.EraseRecorder()
		if e != nil {
			return errors.Wrap(e, "EraseRecorder error")
		}

		/*e = ub.DisconnectFromDevice()
		if e != nil {
			return errors.Wrap(e, "DisconnectFromDevice error")
		}*/
		return nil
	}, ub, t)

	if err != nil {
		t.Errorf("TestReboot error %v\n", err)
	}
}
