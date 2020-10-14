package ubloxbluetooth

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func (ub *UbloxBluetooth) cmdRS232Settings(arg string) (*RS232SettingsReply, error) {
	b, err := ub.writeAndWait(RS232SettingsCommand(arg), true)
	if err != nil {
		return nil, err
	}
	return ProcessRS232SettingsReply(b)
}

// GetRS232Settings allows us to see how the Ublox comms are configured
func (ub *UbloxBluetooth) GetRS232Settings() (*RS232SettingsReply, error) {
	b, err := ub.writeAndWait(RS232SettingsCommand(""), true)
	if err != nil {
		return nil, err
	}
	return ProcessRS232SettingsReply(b)
}

// SetRS232BaudRate - sets the baudrate
func (ub *UbloxBluetooth) SetRS232BaudRate(rate int) error {
	_, err := ub.writeAndWait(RS232SettingsCommand(fmt.Sprintf("%d,1,8,1,1,0", rate)), false)
	if err != nil {
		return err
	}
	return nil
}

// FactoryReset must be called with caution...
func (ub *UbloxBluetooth) FactoryReset() error {
	_, err := ub.writeAndWait(FactoryResetCommand(), false)
	return err
}

// StartMode is an enumerator type
type StartMode byte

// CommandMode 0x00
var CommandMode = StartMode(0x00)

// DataMode 0x01
var DataMode = StartMode(0x01)

// ExtendedDataMode 0x02
var ExtendedDataMode = StartMode(0x02)

// PPPMode 0x03
var PPPMode = StartMode(0x03)

// SetModuleStartMode issues the command to configure the module's start mode
func (ub *UbloxBluetooth) SetModuleStartMode(m StartMode) error {
	d, err := ub.writeAndWait(ModuleStartCommand(m), false)
	if err != nil {
		return err
	}
	fmt.Printf("UMSM: %s [%X]", d, d)
	_, err = ub.writeAndWait(BLEStoreConfig(), false)
	return err
}

// ConfigureUblox setups the ublox module
func (ub *UbloxBluetooth) ConfigureUblox(connectionTimeout time.Duration) error {
	_, err := ub.writeAndWait(BLERole(bleCentral), false)
	if err != nil {
		return errors.Wrap(err, "Error configuring BLECentral")
	}

	_, err = ub.writeAndWait(BLEConfig(maxConnectionInterval, 40), false)
	if err != nil {
		return errors.Wrap(err, "Error configuring maxConnectionInterval")
	}

	timeout := int(connectionTimeout / time.Millisecond)
	_, err = ub.writeAndWait(BLEConfig(connectCreateConnectionTimeout, timeout), false)
	if err != nil {
		return errors.Wrap(err, "Error configuring connectCreateConnectionTimeout")
	}

	_, err = ub.writeAndWait(BLEStoreConfig(), false)
	if err != nil {
		return errors.Wrap(err, "BLEStoreConfig error")
	}
	return nil
}

const inactivityTimeoutType = 1
const inactivityTimeoutValue = 6000
const disconnectResetType = 2
const disconnectResetValue = 1

func (ub *UbloxBluetooth) watchdogConfiguration(itv int, drv int) error {
	_, err := ub.writeAndWait(WatchdogCommand(inactivityTimeoutType, itv), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(WatchdogCommand(disconnectResetType, drv), false)
	if err != nil {
		return err
	}

	return nil
}

// SetWatchdogConfiguration sets the 8power watchdog configuration
func (ub *UbloxBluetooth) SetWatchdogConfiguration() error {
	return ub.watchdogConfiguration(inactivityTimeoutValue, disconnectResetValue)
}

// ResetWatchdogConfiguration zeroes the watchdog configuration
func (ub *UbloxBluetooth) ResetWatchdogConfiguration() error {
	return ub.watchdogConfiguration(0, 0)
}
