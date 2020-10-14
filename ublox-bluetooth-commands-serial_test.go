package ubloxbluetooth

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func setupForSerialTests(t *testing.T, echoOff bool) (*UbloxBluetooth, error) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
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
	if echoOff {
		err = ub.EchoOff()
		if err != nil {
			t.Errorf("EchoOff error %v\n", err)
		}
	}
	err = ub.ATCommand()
	if err != nil {
		return nil, errors.Wrap(err, "AT error")
	}
	return ub, nil
}

func TestReset(t *testing.T) {
	ub, err := setupForSerialTests(t, false)
	ub.serialPort.SetVerbose(true)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}

	err = ub.ResetSerial()
	if err != nil {
		t.Fatalf("ResetSerial error %v\n", err)
	}

	time.Sleep(5 * time.Second)

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}

	ub.Close()
}

func TestDataMode(t *testing.T) {
	ub, err := setupForSerialTests(t, false)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	ub.serialPort.SetVerbose(true)

	err = ub.EnterDataMode()
	if err != nil {
		t.Fatalf("EnterDataMode error %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)

	err = ub.EnterCommandMode()
	if err != nil {
		t.Fatalf("EnterCommandMode error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}
}
