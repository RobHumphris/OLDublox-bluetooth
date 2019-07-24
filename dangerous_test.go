package ubloxbluetooth

import (
	"testing"

	"github.com/8power/ublox-bluetooth/serial"

	"github.com/pkg/errors"
)

func TestReboot(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("CE1A0B7E9D79r", func(t *testing.T) error {
		serial.SetVerbose(true)
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

	err = connectToDevice("CE1A0B7E9D79r", func(t *testing.T) error {
		serial.SetVerbose(true)
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
