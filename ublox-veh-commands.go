package ubloxbluetooth

import (
	"bytes"
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

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(password []byte) (bool, error) {
	if ub.connectedDevice == nil {
		return false, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, append(unlockCommand, password...)), true)
	if err != nil {
		return false, errors.Wrapf(err, "UnlockDevice error")
	}

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

// WriteName sets the device's name
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

var defaultCreditString = uint8ToString(uint8(DefaultCredit))
var halfwayPoint = DefaultCredit

// SendCredits messages the connected device to say that it can accept `credit` number of messages
func (ub *UbloxBluetooth) SendCredits(credit int) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	creditHex := uint8ToString(uint8(credit))
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, creditCommand, creditHex), false)
	return err
}

// DownloadSlotData downloads slot data from
func (ub *UbloxBluetooth) DownloadSlotData(slot int, slotOffset int, dnh DownloadNotificationHandler, dih DownloadIndicationHandler) error {
	commandParameters := fmt.Sprintf("%s%s%s", uint16ToString(uint16(slot)), uint16ToString(uint16(slotOffset)), defaultCreditString)

	expectedSequence := 0
	return ub.downloadData(readSlotDataCommand, commandParameters, readSlotDataReply, func(d []byte) error {
		if d != nil {
			l := len(d)
			sequenceNumber := stringToInt(string(d[l-4 : l]))
			if sequenceNumber != expectedSequence {
				return fmt.Errorf("sequence number: %d expected %d", sequenceNumber, expectedSequence)
			}

			err := dnh(d[:l-4])
			if err != nil {
				return errors.Wrap(err, "download hander error")
			}
			expectedSequence++
		}
		return nil
	}, func(s []byte) error {
		if bytes.HasPrefix(s, readSlotDataReplyBytes) {
			return dih(string(s[4:]))
		}
		return fmt.Errorf("[DownloadSlotData] indication %s does not start with %s", s, readSlotDataReply)
	})
}

// DownloadEventLog requests a number of log records to be downloaded.
func (ub *UbloxBluetooth) DownloadEventLog(startingIndex int, fn DownloadNotificationHandler) error {
	commandParameters := fmt.Sprintf("%s%s", uint16ToString(uint16(startingIndex)), defaultCreditString)
	return ub.downloadData(readEventLogCommand, commandParameters, readEventLogReply, fn, func(d []byte) error {
		if bytes.HasPrefix(d, readEventLogReplyBytes) {
			return nil
		}
		return fmt.Errorf("[DownloadEventLog] indication %s does not start with %s", d, readEventLogReply)
	})
}

func (ub *UbloxBluetooth) downloadData(command []byte, commandParameters string, reply string, dnh DownloadNotificationHandler, dih func([]byte) error) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, command, commandParameters), true)
	if err != nil {
		return errors.Wrap(err, "[downloadData] Command error")
	}

	expected, err := ProcessEventsReply(d, reply)
	if err != nil {
		return errors.Wrap(err, "[downloadData] ProcessEventsReply error")
	}
	return ub.HandleDataDownload(expected, reply, dnh, dih)
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
