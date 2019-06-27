package ubloxbluetooth

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

var (
	unlockCommand         = []byte{0x00}
	versionCommand        = []byte{0x01}
	getTimeCommand        = []byte{0x02}
	readConfigCommand     = []byte{0x03}
	writeConfigCommand    = []byte{0x04}
	readEventLogCommand   = []byte{0x07}
	abortCommand          = []byte{0x09}
	echoCommand           = []byte{0x0B}
	creditCommand         = []byte{0x11}
	rebootCommand         = []byte{0x13}
	recorderInfo          = []byte{0x20}
	readRecorder          = []byte{0x21}
	queryRecorderMetaData = []byte{0x22}
	readRecorderData      = []byte{0x23}
	rssiCommand           = []byte{0x24}
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

// AbortEventLogRead aborts the read
func (ub *UbloxBluetooth) AbortEventLogRead() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, abortCommand), false)
	return err
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
