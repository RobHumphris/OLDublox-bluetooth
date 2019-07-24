package ubloxbluetooth

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

const readRecorderOffset = 16
const readRecorderDataOffset = 8
const MaxMessageLength = 243

var (
	unlockCommand           = []byte{0x00}
	versionCommand          = []byte{0x01}
	getTimeCommand          = []byte{0x02}
	readConfigCommand       = []byte{0x03}
	writeConfigCommand      = []byte{0x04}
	setTimeCommand          = []byte{0x07}
	abortCommand            = []byte{0x09}
	echoCommand             = []byte{0x0B}
	setSettingCommand       = []byte{0x0D}
	getSettingCommand       = []byte{0x0E}
	creditCommand           = []byte{0x11}
	recorderEraseCommand    = []byte{0x12}
	rebootCommand           = []byte{0x13}
	messageCommand          = []byte{0x14}
	recorderInfoCommand     = []byte{0x20}
	readRecorderCommand     = []byte{0x21}
	queryRecorderCommand    = []byte{0x22}
	readRecorderDataCommand = []byte{0x23}
	rssiCommand             = []byte{0x24}
)

func (ub *UbloxBluetooth) newCharacteristicCommand(handle int, data []byte) characteristicCommand {
	return characteristicCommand{
		ub.connectedDevice.Handle,
		handle,
		data,
	}
}

func (ub *UbloxBluetooth) newCharacteristicHexCommand(handle int, data []byte, hex string) characteristicHexCommand {
	return characteristicHexCommand{
		&characteristicCommand{ub.connectedDevice.Handle, handle, data},
		hex,
	}
}

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(password []byte) (bool, error) {
	if ub.connectedDevice == nil {
		return false, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, append(unlockCommand, password...))
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return false, errors.Wrapf(err, "UnlockDevice error")
	}

	return ProcessUnlockReply(d)
}

// GetVersion request the connected device's version
func (ub *UbloxBluetooth) GetVersion() (*VersionReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, versionCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetVersion error")
	}
	return NewVersionReply(d)
}

// GetTime requests the current device info.
func (ub *UbloxBluetooth) GetTime() (int32, error) {
	if ub.connectedDevice == nil {
		return -1, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, getTimeCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return -1, errors.Wrapf(err, "GetTime error")
	}

	t, err := splitOutResponse(d, infoReply)
	if err != nil {
		return -1, err
	}

	// I hate casting something that has just been cast.
	return int32(stringToInt(t[4:12])), nil
}

// SetTime sets the current time for the device.
func (ub *UbloxBluetooth) SetTime(timestamp int32) (*TimeAdjustReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	tsHex := uint32ToString(uint32(timestamp))
	c := ub.newCharacteristicHexCommand(commandValueHandle, setTimeCommand, tsHex)
	d, err := ub.writeAndWait(writeCharacteristicHexCommand(c), true)
	if err != nil {
		return nil, errors.Wrapf(err, "SetTime error")
	}
	return NewTimeAdjustReply(d)
}

// ReadConfig requests the device's current config
func (ub *UbloxBluetooth) ReadConfig() (*ConfigReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, readConfigCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadConfig error")
	}
	return NewConfigReply(d)
}

// WriteConfig sends the passed config to the device
func (ub *UbloxBluetooth) WriteConfig(cfg *ConfigReply) error {
	if ub.connectedDevice == nil {
		return ErrNotConnected
	}

	c := ub.newCharacteristicHexCommand(commandValueHandle, writeConfigCommand, cfg.ByteArray())
	_, err := ub.writeAndWait(writeCharacteristicHexCommand(c), true)
	return err
}

// DefaultCredit says that we can handle 16 messages in our FIFO
const DefaultCredit = 16

var defaultCreditString = uint8ToString(uint8(DefaultCredit))
var halfwayPoint = DefaultCredit

// SendCredits messages the connected device to say that it can accept `credit` number of messages
func (ub *UbloxBluetooth) SendCredits(credit int) error {
	if ub.connectedDevice == nil {
		return ErrNotConnected
	}

	creditHex := uint8ToString(uint8(credit))
	c := ub.newCharacteristicHexCommand(commandValueHandle, creditCommand, creditHex)
	_, err := ub.writeAndWait(writeCharacteristicHexCommand(c), false)
	return err
}

// EraseRecorder issues the erase command - which wipes the sensor (use with care)
func (ub *UbloxBluetooth) EraseRecorder() error {
	if ub.connectedDevice == nil {
		return ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, recorderEraseCommand)
	_, err := ub.writeAndWait(writeCharacteristicCommand(c), false)
	return err
}

func (ub *UbloxBluetooth) simpleCommand(cmd []byte) error {
	if ub.connectedDevice == nil {
		return ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, cmd)
	_, err := ub.writeAndWait(writeCharacteristicCommand(c), false)
	return err
}

// RebootRecorder issues the reboot command to the sensor.
func (ub *UbloxBluetooth) RebootRecorder() error {
	return ub.simpleCommand(rebootCommand)
}

// AbortEventLogRead aborts the read
func (ub *UbloxBluetooth) AbortEventLogRead() error {
	return ub.simpleCommand(abortCommand)
}

// WriteMessage writes `msg` string to the device's event log. messageCommand
func (ub *UbloxBluetooth) WriteMessage(msg string) error {
	if ub.connectedDevice == nil {
		return ErrNotConnected
	}

	msgLen := len(msg)
	if msgLen > MaxMessageLength {
		msgLen = MaxMessageLength
	}

	c := ub.newCharacteristicHexCommand(commandValueHandle, writeConfigCommand, stringToHexString(msg[:msgLen]))
	_, err := ub.writeAndWait(writeCharacteristicHexCommand(c), false)
	return err
}

// EchoCommand sends the `data` string as bytes, and receives something in return.
func (ub *UbloxBluetooth) EchoCommand(data string) (bool, error) {
	if ub.connectedDevice == nil {
		return false, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, echoCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return false, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessEchoReply(d)
}

// ReadRecorderInfo reads the Recorder information
func (ub *UbloxBluetooth) ReadRecorderInfo() (*RecorderInfoReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, recorderInfoCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return nil, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessReadRecorderInfoReply(d)
}

// ReadRecorder downloads the record entries starting from the given sequence.
// Each response is converted to a VehEvent and the function `fn` is invoked with it.
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
		return ErrNotConnected
	}

	c := ub.newCharacteristicHexCommand(commandValueHandle, command, commandParameters)
	d, err := ub.writeAndWait(writeCharacteristicHexCommand(c), true)
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
func (ub *UbloxBluetooth) QueryRecorderMetaDataCommand(sequence uint32) (*RecorderMetaDataReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	cmd := make([]byte, 5)
	cmd[0] = queryRecorderCommand[0]
	binary.LittleEndian.PutUint32(cmd[1:], uint32(sequence))

	c := ub.newCharacteristicCommand(commandValueHandle, cmd)

	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return nil, errors.Wrap(err, "RecorderInfo error")
	}
	return ProcessQueryRecorderMetaDataReply(d)
}

// ReadRecorderDataCommand issues the readRecorderDataCommand and handles the onslaught of data thats returned
func (ub *UbloxBluetooth) ReadRecorderDataCommand(sequence uint32, md *RecorderMetaDataReply) ([]byte, error) {
	data := []byte{}
	commandParameters := fmt.Sprintf("%s%s", uint32ToString(sequence), defaultCreditString)
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

// GetRSSI returns the RSSI value for the gateway from the connected device
func (ub *UbloxBluetooth) GetRSSI() (*RSSIReply, error) {
	if ub.connectedDevice == nil {
		return nil, ErrNotConnected
	}

	c := ub.newCharacteristicCommand(commandValueHandle, rssiCommand)
	d, err := ub.writeAndWait(writeCharacteristicCommand(c), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetRSSI error")
	}
	return NewRSSIReply(d)
}
