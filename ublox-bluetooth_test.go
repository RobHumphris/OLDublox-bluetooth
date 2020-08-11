package ubloxbluetooth

import (
	"fmt"
	"testing"
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
