package ubloxbluetooth

import (
	"time"

	"github.com/pkg/errors"
)

// Section 5.1.3 on page 30 of
// https://www.u-blox.com/sites/default/files/u-connect-ATCommands-Manual_%28UBX-14044127%29.pdf
// states that a delay of 50ms is required before start of data transmission.
func modeSwitchDelay() {
	time.Sleep(50 * time.Millisecond)
}

// EnterDataMode sends the ATO command to set Ublox to Data Mode
func (ub *UbloxBluetooth) EnterDataMode() error {
	err := ub.Write(EnterDataModeCommand().Cmd)
	if err != nil {
		return errors.Wrap(err, "[EnterDataMode] error")
	}
	ub.currentMode = dataMode
	ub.serialPort.SetEDMFlag(true)
	modeSwitchDelay()
	return nil
}

// EnterExtendedDataMode sends the ATO2 command to set Ublox to
// Extended Data Mode (EDM)
func (ub *UbloxBluetooth) EnterExtendedDataMode() error {
	err := ub.Write(EnterExtendedDataModeCommand().Cmd)
	if err != nil {
		return errors.Wrap(err, "[EnterExtendedDataMode] error")
	}
	ub.currentMode = extendedDataMode
	ub.serialPort.SetEDMFlag(true)
	modeSwitchDelay()
	return nil
}

// EnterCommandMode sends the Escape Sequence required to return the Command Mode (AT)
func (ub *UbloxBluetooth) EnterCommandMode() error {
	err := ub.serialPort.ToggleDTR()
	if err != nil {
		return errors.Wrap(err, "[EnterCommandMode] error")
	}
	ub.currentMode = dataMode
	modeSwitchDelay()
	return nil
}
