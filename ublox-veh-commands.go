package ubloxbluetooth

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

var (
	unlockCommand        = []byte{0x00}
	versionCommand       = []byte{0x01}
	getTimeCommand       = []byte{0x02}
	readConfigCommand    = []byte{0x03}
	writeConfigCommand   = []byte{0x04}
	readNameCommand      = []byte{0x05}
	writeNameCommand     = []byte{0x06}
	readEventLogCommand  = []byte{0x07}
	clearEventLogCommand = []byte{0x08}
	abortCommand         = []byte{0x09}
	readSlotCountCommand = []byte{0x0E}
	readSlotInfoCommand  = []byte{0x0F}
	readSlotDataCommand  = []byte{0x10}
	creditCommand        = []byte{0x11}
	eraseSlotCommand     = []byte{0x12}
	rebootCommand        = []byte{0x13}

	// Version 2 commands
	recorderInfo          = []byte{0x20}
	readRecorder          = []byte{0x21}
	queryRecorderMetaData = []byte{0x22}
	readRecorderData      = []byte{0x23}
)

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

// GetTime requests the current device info.
func (ub *UbloxBluetooth) GetTime() (int, error) {
	if ub.connectedDevice == nil {
		return -1, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, getTimeCommand), true)
	if err != nil {
		return -1, errors.Wrapf(err, "GetInfo error")
	}

	t, err := splitOutResponse(d, infoReply)
	if err != nil {
		return -1, err
	}

	return stringToInt(t[4:12]), nil
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

	configData := cfg.ByteArray()
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, writeConfigCommand, configData), true)
	return err
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

	return NewNameReply(d)
}

// WriteName sets the device's name
func (ub *UbloxBluetooth) WriteName(name string) error {
	stringBytes := fmt.Sprintf("%x", name)

	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, writeNameCommand, stringBytes), true)
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

// EraseSlotData requests that the device erases its slots...
func (ub *UbloxBluetooth) EraseSlotData() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, eraseSlotCommand), true)
	if err != nil {
		return errors.Wrap(err, "EraseSlotData error")
	}
	return ProcessEraseSlotDataReply(d)
}

func (ub *UbloxBluetooth) ReadRecorderInfo() (*RecorderInfoReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, recorderInfo), true)
	if err != nil {
		return nil, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessReadRecorderInfoReply(d)
}
