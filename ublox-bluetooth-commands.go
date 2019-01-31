package ubloxbluetooth

import (
	"bytes"
	"fmt"
)

// ATCommand issues a straight AT command - used to test connection
func (ub *UbloxBluetooth) ATCommand() error {
	err := ub.Write(ATCommand())
	if err != nil {
		return err
	}
	_, err = ub.WaitForResponse(false)
	return err
}

// RebootUblox reboots the Ublox chip
func (ub *UbloxBluetooth) RebootUblox() error {
	err := ub.Write(RebootCommand())
	if err != nil {
		return err
	}

	b, err := ub.WaitForResponse(true)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, rebootResponse) {
		return fmt.Errorf("unexpected reboot response %s", b)
	}
	return nil
}

// DiscoveryCommand issues the Discover command and builds a list of new devices
func (ub *UbloxBluetooth) DiscoveryCommand() ([]DiscoveryReply, error) {
	err := ub.Write(DiscoveryCommand())
	if err != nil {
		return nil, err
	}

	d, err := ub.WaitForResponse(true)
	if err != nil {
		return nil, err
	}
	return ProcessDiscoveryReply(d)
}

// ConnectToDevice attempts to connect to the device with the specified address.
func (ub *UbloxBluetooth) ConnectToDevice(addr string) (*ConnectionReply, error) {
	err := ub.Write(ConnectCommand(addr))
	if err != nil {
		return nil, err
	}

	d, err := ub.WaitForResponse(true)
	if err != nil {
		return nil, err
	}
	return NewConnectionReply(string(d))
}

// DisconnectFromDevice issues the disconnect command using the handle from the ConnectionReply
func (ub *UbloxBluetooth) DisconnectFromDevice(cr *ConnectionReply) error {
	err := ub.Write(DisconnectCommand(cr.Handle))
	if err != nil {
		return err
	}

	_, err = ub.WaitForResponse(false)
	return err
}

// EnableIndications instructs the connected device to initialise indiciations
func (ub *UbloxBluetooth) EnableIndications(cr *ConnectionReply) error {
	err := ub.Write(WriteCharacteristicConfigurationCommand(cr.Handle, commandCCCDHandle, 2))
	if err != nil {
		return err
	}
	_, err = ub.WaitForResponse(false)
	return err
}

// EnableNotifications instructs the connected device to initialise notifications
func (ub *UbloxBluetooth) EnableNotifications(cr *ConnectionReply) error {
	err := ub.Write(WriteCharacteristicConfigurationCommand(cr.Handle, dataCCCDHandle, 1))
	if err != nil {
		return err
	}
	_, err = ub.WaitForResponse(false)
	return err
}

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(cr *ConnectionReply, password []byte) (bool, error) {
	cmd := append(unlockCommand, password...)
	err := ub.Write(WriteCharacteristicCommand(cr.Handle, commandValueHandle, cmd))
	if err != nil {
		return false, err
	}

	d, err := ub.WaitForResponse(true)
	if err != nil {
		return false, err
	}
	return ProcessUnlockReply(d)
}

// GetVersion request the connected device's version
func (ub *UbloxBluetooth) GetVersion(cr *ConnectionReply) (*VersionReply, error) {
	err := ub.Write(WriteCharacteristicCommand(cr.Handle, commandValueHandle, versionCommand))
	if err != nil {
		return nil, err
	}

	d, err := ub.WaitForResponse(true)
	if err != nil {
		return nil, err
	}
	return NewVersionReply(d)
}

// GetInfo requests the current device info.
func (ub *UbloxBluetooth) GetInfo(cr *ConnectionReply) (*InfoReply, error) {
	err := ub.Write(WriteCharacteristicCommand(cr.Handle, commandValueHandle, infoCommand))
	if err != nil {
		return nil, err
	}
	d, err := ub.WaitForResponse(true)
	if err != nil {
		return nil, err
	}
	return NewInfoReply(d)
}

// ReadConfig requests the device's current config
func (ub *UbloxBluetooth) ReadConfig(cr *ConnectionReply) (*ConfigReply, error) {
	err := ub.Write(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readConfigCommand))
	if err != nil {
		return nil, err
	}
	d, err := ub.WaitForResponse(true)
	if err != nil {
		return nil, err
	}
	return NewConfigReply(d)
}

/*func (ub *UbloxBluetooth) DownloadLogFile(cr ConnectionReply) error {
	err := ub.Write(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readEventLogCommand
}*/
