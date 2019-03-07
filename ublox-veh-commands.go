package ubloxbluetooth

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

var unlockCommand = []byte{0x00}
var versionCommand = []byte{0x01}
var infoCommand = []byte{0x02}
var readConfigCommand = []byte{0x03}
var writeConfigCommand = []byte{0x04}
var readNameCommand = []byte{0x05}
var writeNameCommand = []byte{0x06}
var readEventLogCommand = []byte{0x07}
var clearEventLogCommand = []byte{0x08}
var abortCommand = []byte{0x09}
var readSlotCountCommand = []byte{0x0E}
var readSlotInfoCommand = []byte{0x0F}
var readSlotDataCommand = []byte{0x10}
var creditCommand = []byte{0x11}

// DiscoveryCommand issues the Discover command and calls the DiscoveryReplyHandler
func (ub *UbloxBluetooth) DiscoveryCommand(fn DiscoveryReplyHandler) error {
	dc := DiscoveryCommand()
	err := ub.Write(dc.Cmd)
	if err != nil {
		return err
	}

	return ub.HandleDiscovery(dc.Resp, func(d []byte) (bool, error) {
		dr, err := ProcessDiscoveryReply(d)
		if err == nil {
			err = fn(dr)
		} else if err != ErrUnexpectedResponse {
			return false, err
		}
		return true, nil
	})
}

// ConnectToDevice attempts to connect to the device with the specified address.
func (ub *UbloxBluetooth) ConnectToDevice(addr string, onConnect DeviceEvent, onDisconnect DeviceEvent) error {
	d, err := ub.writeAndWait(ConnectCommand(addr), true)
	if err != nil {
		return err
	}

	cr, err := NewConnectionReply(string(d))
	if err != nil {
		return err
	}

	ub.connectedDevice = cr
	ub.disconnectHandler = onDisconnect
	ub.disconnectExpected = false
	return onConnect()
}

func (ub *UbloxBluetooth) handleUnexpectedDisconnection() {
	ub.connectedDevice = nil
	ub.disconnectHandler = nil
	ub.disconnectExpected = false
	if ub.disconnectHandler != nil {
		ub.ErrorChannel <- ub.disconnectHandler()
	}
}

// DisconnectFromDevice issues the disconnect command using the handle from the ConnectionReply
func (ub *UbloxBluetooth) DisconnectFromDevice() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	ub.disconnectExpected = true

	d, err := ub.writeAndWait(DisconnectCommand(ub.connectedDevice.Handle), true)
	if err != nil {
		return err
	}

	ok, err := ProcessDisconnectReply(d)
	if !ok {
		return fmt.Errorf("Incorrect disconnect reply %q", d)
	}
	ub.connectedDevice = nil
	ub.disconnectHandler = nil
	ub.disconnectExpected = false
	return err
}

// EnableIndications instructs the connected device to initialise indiciations
func (ub *UbloxBluetooth) EnableIndications() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicConfigurationCommand(ub.connectedDevice.Handle, commandCCCDHandle, 2), false)
	return err
}

// EnableNotifications instructs the connected device to initialise notifications
func (ub *UbloxBluetooth) EnableNotifications() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicConfigurationCommand(ub.connectedDevice.Handle, dataCCCDHandle, 1), false)
	return err
}

// ReadCharacterisitic reads the
func (ub *UbloxBluetooth) ReadCharacterisitic() ([]byte, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}
	d, err := ub.writeAndWait(ReadCharacterisiticCommand(ub.connectedDevice.Handle, commandValueHandle), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadCharacterisitic error")
	}
	fmt.Printf("ReadCharacterisitic: %s\n", d)
	return d, nil
}

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(password []byte) (bool, error) {
	if ub.connectedDevice == nil {
		return false, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, append(unlockCommand, password...)), true)
	if err != nil {
		return false, errors.Wrapf(err, "UnlockDevice error")
	}

	ub.ReadCharacterisitic()

	return ProcessUnlockReply(d)
}

// GetVersion request the connected device's version
func (ub *UbloxBluetooth) GetVersion() (*VersionReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, versionCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetVersion error")
	}
	return NewVersionReply(d)
}

// GetInfo requests the current device info.
func (ub *UbloxBluetooth) GetInfo() (*InfoReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, infoCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetInfo error")
	}

	ub.ReadCharacterisitic()

	return NewInfoReply(d)
}

// ReadConfig requests the device's current config
func (ub *UbloxBluetooth) ReadConfig() (*ConfigReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, readConfigCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadConfig error")
	}
	return NewConfigReply(d)
}

// WriteConfig sends the passed config to the device
func (ub *UbloxBluetooth) WriteConfig(cfg *ConfigReply) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	configData, err := cfg.ByteArray()
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, writeConfigCommand, configData), true)
	return fmt.Errorf("NOT IMPLEMENTED")
}

// ReadName messages the remote device to get its set name
func (ub *UbloxBluetooth) ReadName() (string, error) {
	name := ""
	if ub.connectedDevice == nil {
		return name, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, readNameCommand), true)
	if err != nil {
		return name, errors.Wrapf(err, "readNameCommand error")
	}

	name = string(d)

	return name, nil
}

func (ub *UbloxBluetooth) WriteName(name string) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, writeNameCommand, name), true)
	if err != nil {
		return errors.Wrapf(err, "writeNameCommand error")
	}
	return nil
}

// DefaultCredit says that we can handle 16 messages in our FIFO
const DefaultCredit = 16

var halfwayPoint = DefaultCredit / 2

// SendCredits messages the connected device to say that it can accept `credit` number of messages
func (ub *UbloxBluetooth) SendCredits(credit int) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	creditHex := uint8ToString(uint8(credit))
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, creditCommand, creditHex), false)
	return err
}

// DownloadLogFile requests a number of log records to be downloaded.
func (ub *UbloxBluetooth) DownloadLogFile(startingIndex int, fn DownloadLogHandler) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}
	si := uint16ToString(uint16(startingIndex))

	d, errr := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, readEventLogCommand, si), true)
	if errr != nil {
		return errors.Wrap(errr, "readEventLogCommand error")
	}

	expected, errr := ProcessEventsReply(d)
	if errr != nil {
		return errors.Wrap(errr, "ProcessEventsReply error")
	}

	errr = ub.SendCredits(DefaultCredit)
	if errr != nil {
		return errors.Wrap(errr, "SendCredits error")
	}

	notificationsReceived := 0

	received, errr := ub.HandleDataDownload(expected, func(d []byte) (bool, error) {
		var err error
		if bytes.HasPrefix(d, gattNotificationResponse) {
			d, e := splitOutNotification(d, readEventLogReply)
			if e != nil {
				err = errors.Wrapf(err, e.Error())
			} else {
				notificationsReceived++
				dt, e := hex.DecodeString(string(d[:]))
				if e != nil {
					err = errors.Wrapf(err, e.Error())
				} else {
					e = fn(dt)
					if e != nil {
						err = errors.Wrapf(err, e.Error())
					}
				}
				if notificationsReceived%halfwayPoint == 0 {
					e = ub.SendCredits(halfwayPoint)
					if e != nil {
						err = errors.Wrapf(err, e.Error())
					}
				}
			}
		} else if bytes.HasPrefix(d, gattIndicationResponse) {
			_, e := splitOutResponse(d, readEventLogReply)
			if err == nil {
				err = e
			} else {
				err = errors.Wrapf(err, "notification error %v ", e)
			}
			return false, e
		}
		return true, err
	})

	if received != expected {
		errr = errors.Wrap(errr, fmt.Sprintf("expected %d received %d\n", expected, received))
	}
	return errr
}

// ClearEventLog requests that the event log of the connected device be cleared.
func (ub *UbloxBluetooth) ClearEventLog() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, clearEventLogCommand), true)
	if err != nil {
		return errors.Wrap(err, "ClearEventLog error")
	}
	return ProcessClearEventReply(d)
}

// AbortEventLogRead aborts the read
func (ub *UbloxBluetooth) AbortEventLogRead() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, abortCommand), false)
	return err
}

// ReadSlotCount get recorder slot count
func (ub *UbloxBluetooth) ReadSlotCount() (*SlotCountReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, readSlotCountCommand), true)
	if err != nil {
		return nil, errors.Wrap(err, "ReadSlotCount error")
	}
	return NewSlotCountReply(d)
}

// ReadSlotInfo get recorder's slot info for the provided slotNumber, returns a SlotInfoReply structure or an error
func (ub *UbloxBluetooth) ReadSlotInfo(slotNumber int) (*SlotInfoReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	slot := uint16ToString(uint16(slotNumber))
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, readSlotInfoCommand, slot), true)
	if err != nil {
		return nil, err
	}
	return NewSlotInfoReply(d)
}

// ReadSlotData gets the data for the given slot and offset
func (ub *UbloxBluetooth) ReadSlotData(slotNumber int, offset int, requiredBytes int) ([]byte, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	slot := uint16ToString(uint16(slotNumber))
	off := uint16ToString(uint16(offset))
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, readSlotDataCommand, slot+off), true)
	if err != nil {
		return nil, err
	}

	expectedNotifications, err := ProcessSlotsReply(d)
	actualNotifications := 0
	data := []byte{}
	_, err = ub.HandleDataDownload(expectedNotifications, func(d []byte) (bool, error) {
		if bytes.HasPrefix(d, gattNotificationResponse) {
			actualNotifications++
			d, e := splitOutNotification(d, readSlotInfoReply)
			if e != nil {
				return false, e
			}
			bytes, e := hex.DecodeString(string(d[:]))
			if err != nil {
				return false, err
			}
			data = append(data, bytes...)
			if len(data) < requiredBytes {
				return true, nil
			}
		}
		return false, nil
	})

	return nil, err
}
