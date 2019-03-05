package ubloxbluetooth

import (
	"log"
	"testing"

	serial "github.com/RobHumphris/ublox-bluetooth/serial"
)

var bt *UbloxBluetooth

func handleFatal(s string, err error) {
	bt.Close()
	log.Fatalf("%s %v\n", s, err)
}

func TestRestart(t *testing.T) {
	var err error
	serial.SetVerbose(true)

	bt, err = NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		handleFatal("NewUbloxBluetooth error", err)
	}
	defer bt.Close()

	err = bt.ATCommand()
	if err != nil {
		handleFatal("AT - 0 error", err)
	}

	err = bt.EnterExtendedDataMode()
	if err != nil {
		handleFatal("EnterExtendedDataMode error", err)
	}

	err = bt.ATCommand()
	if err != nil {
		handleFatal("AT - 1 error", err)
	}

	err = bt.RebootUblox()
	if err != nil {
		handleFatal("RebootUblox error", err)
	}

	err = bt.EnterExtendedDataMode()
	if err != nil {
		handleFatal("EnterExtendedDataMode error", err)
	}

	err = bt.ATCommand()
	if err != nil {
		handleFatal("AT - 2 error", err)
	}
}
