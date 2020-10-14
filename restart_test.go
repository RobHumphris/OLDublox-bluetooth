package ubloxbluetooth

import (
	"fmt"
	"log"
	"testing"
)

var bt *UbloxBluetooth

func testErrorHandler(err error) {
	fmt.Printf("genericErrorHandler: %v\n", err)
}

func handleFatal(s string, err error) {
	bt.Close()
	log.Fatalf("%s %v\n", s, err)
}

func TestResetWatchdog(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer bt.Close()

	err = ub.ATCommand()
	if err != nil {
		handleFatal("AT - 0 error", err)
	}

	err = ub.ResetWatchdogConfiguration()
	if err != nil {
		fmt.Printf("ResetWatchdogConfiguration error %v\n", err)
	}
	fmt.Println("ResetWatchdogConfiguration OK")
}

func TestSetWatchdog(t *testing.T) {
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer ub.Close()

	err = ub.ATCommand()
	if err != nil {
		handleFatal("AT - 0 error", err)
	}

	err = ub.SetWatchdogConfiguration()
	if err != nil {
		fmt.Printf("ResetWatchdogConfiguration error %v\n", err)
	}
	fmt.Println("ResetWatchdogConfiguration OK")
}

func TestRestartViaDTR(t *testing.T) {
	var err error

	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer bt.Close()
	ub.serialPort.SetVerbose(true)

	ub.ResetUblox()

	err = ub.ATCommand()
	if err != nil {
		handleFatal("AT - 0 error", err)
	}
}

func TestRestart(t *testing.T) {
	var err error
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	defer bt.Close()
	ub.serialPort.SetVerbose(true)

	err = ub.ATCommand()
	if err != nil {
		handleFatal("AT - 0 error", err)
	}

	/*err = bt.EnterExtendedDataMode()
	if err != nil {
		handleFatal("EnterExtendedDataMode error", err)
	}*/

	err = ub.ATCommand()
	if err != nil {
		handleFatal("AT - 1 error", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		handleFatal("RebootUblox error", err)
	}

	err = ub.EnterExtendedDataMode()
	if err != nil {
		handleFatal("EnterExtendedDataMode error", err)
	}

	err = ub.MultipleATCommands()
	if err != nil {
		handleFatal("AT - 2 error", err)
	}
}
