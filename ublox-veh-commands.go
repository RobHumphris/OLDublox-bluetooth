package ubloxbluetooth

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

const readRecorderOffset = 16
const readRecorderDataOffset = 8

var (
	unlockCommand           = []byte{0x00}
	versionCommand          = []byte{0x01}
	getTimeCommand          = []byte{0x02}
	readConfigCommand       = []byte{0x03}
	writeConfigCommand      = []byte{0x04}
	abortCommand            = []byte{0x09}
	echoCommand             = []byte{0x0B}
	creditCommand           = []byte{0x11}
	rebootCommand           = []byte{0x13}
	recorderInfoCommand     = []byte{0x20}
	readRecorderCommand     = []byte{0x21}
	queryRecorderCommand    = []byte{0x22}
	readRecorderDataCommand = []byte{0x23}
	rssiCommand             = []byte{0x24}
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

// AbortEventLogRead aborts the read
func (ub *UbloxBluetooth) AbortEventLogRead() error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, abortCommand), false)
	return err
}

// EchoCommand sends the `data` string as bytes, and receives something in return.
func (ub *UbloxBluetooth) EchoCommand(data string) (bool, error) {
	if ub.connectedDevice == nil {
		return false, fmt.Errorf("ConnectionReply is nil")
	}
	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, echoCommand), true)
	if err != nil {
		return false, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessEchoReply(d)
}

// ReadRecorderInfo reads the Recorder information
func (ub *UbloxBluetooth) ReadRecorderInfo() (*RecorderInfoReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, recorderInfoCommand), true)
	if err != nil {
		return nil, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessReadRecorderInfoReply(d)
}

// ReadRecorder downloads the record entries for the given `sequence`
func (ub *UbloxBluetooth) ReadRecorder(sequence uint32, fn func(*VehEvent) error) error {
	commandParameters := fmt.Sprintf("%s%s", uint32ToString(sequence), defaultCreditString)
	err := ub.downloadData(readRecorderCommand, commandParameters, readRecorderOffset, readRecorderReply, func(d []byte) error {
		if d != nil {
			b, err := hex.DecodeString(string(d))
			if err != nil {
				return err
			}
			ve, err := NewRecorderEvent(b)
			if err == nil {
				return fn(ve)
			}
		}
		return nil
	}, func(d []byte) error {
		return nil
	})
	return err
}

func (ub *UbloxBluetooth) downloadData(command []byte, commandParameters string, lengthOffset int, reply string, dnh func([]byte) error, dih func([]byte) error) error {
	if ub.connectedDevice == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(ub.connectedDevice.Handle, commandValueHandle, command, commandParameters), true)
	if err != nil {
		return errors.Wrap(err, "[downloadData] Command error")
	}

	t, err := splitOutResponse(d, reply)
	if err != nil {
		return errors.Wrap(err, "[downloadData] processEventsReply error")
	}
	return ub.HandleDataDownload(stringToInt(t[lengthOffset-4:]), reply, dnh, dih)
}

// QueryRecorderMetaDataCommand gets the
func (ub *UbloxBluetooth) QueryRecorderMetaDataCommand(sequence int) (*RecorderMetaDataReply, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	cmd := make([]byte, 5)
	cmd[0] = queryRecorderCommand[0]
	binary.LittleEndian.PutUint32(cmd[1:], uint32(sequence))

	d, err := ub.writeAndWait(WriteCharacteristicCommand(ub.connectedDevice.Handle, commandValueHandle, cmd), true)
	if err != nil {
		return nil, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessQueryRecorderMetaDataReply(d)
}

// ReadRecorderDataCommand issues the readRecorderDataCommand and handles the onslaught of data thats returned
func (ub *UbloxBluetooth) ReadRecorderDataCommand(sequence int, md *RecorderMetaDataReply) ([]byte, error) {
	data := []byte{}
	commandParameters := fmt.Sprintf("%s%s", uint32ToString(uint32(sequence)), defaultCreditString)
	err := ub.downloadData(readRecorderDataCommand, commandParameters, readRecorderDataOffset, readRecorderDataReply, func(d []byte) error {
		if d != nil {
			b, err := hex.DecodeString(string(d))
			if err != nil {
				return err
			}
			data = append(data, b...)
		}
		return nil
	}, func(d []byte) error {
		return nil
	})
	return data, err
}
