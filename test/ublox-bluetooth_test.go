package ubloxbluetooth

import (
	"fmt"
	"testing"

	u "github.com/RobHumphris/ublox-bluetooth"
	serial "github.com/RobHumphris/ublox-bluetooth/serial"
)

func TestATCommand(t *testing.T) {
	ub, err := u.NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	serial.SetVerbose(true)

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
