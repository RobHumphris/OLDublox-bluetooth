package ubloxbluetooth

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

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

func (ub *UbloxBluetooth) writeAndWait(r CmdResp, waitForData bool) ([]byte, error) {
	err := ub.Write(r.Cmd)
	if err != nil {
		return nil, err
	}
	return ub.WaitForResponse(r.Resp, waitForData)
}

// ATCommand issues a straight AT command - used to test connection
func (ub *UbloxBluetooth) ATCommand() error {
	_, err := ub.writeAndWait(ATCommand(), false)
	return err
}

// MultipleATCommands sends upto 5 AT commands - used to ensure stable connection.
func (ub *UbloxBluetooth) MultipleATCommands() error {
	var e error
	for i := 0; i < 5; i++ {
		time.Sleep(50 * time.Millisecond)
		err := ub.ATCommand()
		if err == nil {
			return nil
		}
		e = errors.Wrapf(e, "AT Command error %v", err)
	}
	return fmt.Errorf("Failed after 5 attempts %v", e)
}

// EchoOff requests that the ublox device is a little less noisy
func (ub *UbloxBluetooth) EchoOff() error {
	_, err := ub.writeAndWait(EchoOffCommand(), false)
	return err
}

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
	d, err := ub.writeAndWait(ModuleStartCommand(m), true)
	if err != nil {
		return err
	}
	fmt.Printf("UMSM: %s [%X]", d, d)
	return nil
}

// ConfigureUblox setups the ublox module
func (ub *UbloxBluetooth) ConfigureUblox() error {
	_, err := ub.writeAndWait(BLERole(bleCentral), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEConfig(minConnectionInterval, 24), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEConfig(maxConnectionInterval, 40), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEStoreConfig(), false)
	return err
}

// RebootUblox reboots the Ublox chip
func (ub *UbloxBluetooth) RebootUblox() error {
	r := RebootCommand()
	err := ub.Write(r.Cmd)
	if err != nil {
		return err
	}
	ub.currentMode = dataMode
	_, err = ub.WaitForResponse(r.Resp, false)
	modeSwitchDelay()
	return err
}

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
	return onConnect(cr)
}

// DisconnectFromDevice issues the disconnect command using the handle from the ConnectionReply
func (ub *UbloxBluetooth) DisconnectFromDevice(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	ub.disconnectExpected = true

	d, err := ub.writeAndWait(DisconnectCommand(cr.Handle), true)
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
func (ub *UbloxBluetooth) EnableIndications(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicConfigurationCommand(cr.Handle, commandCCCDHandle, 2), false)
	return err
}

// EnableNotifications instructs the connected device to initialise notifications
func (ub *UbloxBluetooth) EnableNotifications(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicConfigurationCommand(cr.Handle, dataCCCDHandle, 1), false)
	return err
}

// ReadCharacterisitic reads the
func (ub *UbloxBluetooth) ReadCharacterisitic(cr *ConnectionReply) ([]byte, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}
	d, err := ub.writeAndWait(ReadCharacterisiticCommand(cr.Handle, commandValueHandle), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadCharacterisitic error")
	}
	fmt.Printf("ReadCharacterisitic: %s\n", d)
	return d, nil
}

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(cr *ConnectionReply, password []byte) (bool, error) {
	if cr == nil {
		return false, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, append(unlockCommand, password...)), true)
	if err != nil {
		return false, errors.Wrapf(err, "UnlockDevice error")
	}

	ub.ReadCharacterisitic(cr)

	return ProcessUnlockReply(d)
}

// GetVersion request the connected device's version
func (ub *UbloxBluetooth) GetVersion(cr *ConnectionReply) (*VersionReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, versionCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetVersion error")
	}
	return NewVersionReply(d)
}

// GetInfo requests the current device info.
func (ub *UbloxBluetooth) GetInfo(cr *ConnectionReply) (*InfoReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, infoCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "GetInfo error")
	}

	ub.ReadCharacterisitic(cr)

	return NewInfoReply(d)
}

// ReadConfig requests the device's current config
func (ub *UbloxBluetooth) ReadConfig(cr *ConnectionReply) (*ConfigReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readConfigCommand), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadConfig error")
	}
	return NewConfigReply(d)
}

// WriteConfig sends the passed config to the device
func (ub *UbloxBluetooth) WriteConfig(cfg *ConfigReply, cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	configData, err := cfg.ByteArray()
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, writeConfigCommand, configData), true)
	return fmt.Errorf("NOT IMPLEMENTED")
}

// ReadName messages the remote device to get its set name
func (ub *UbloxBluetooth) ReadName(cr *ConnectionReply) (string, error) {
	name := ""
	if cr == nil {
		return name, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readNameCommand), true)
	if err != nil {
		return name, errors.Wrapf(err, "readNameCommand error")
	}

	name = string(d)

	return name, nil
}

func (ub *UbloxBluetooth) WriteName(name string, cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, writeNameCommand, name), true)
	if err != nil {
		return errors.Wrapf(err, "writeNameCommand error")
	}
	return nil
}

// DefaultCredit says that we can handle 16 messages in our FIFO
const DefaultCredit = 16

var halfwayPoint = DefaultCredit / 2

// SendCredits messages the connected device to say that it can accept `credit` number of messages
func (ub *UbloxBluetooth) SendCredits(cr *ConnectionReply, credit int) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	creditHex := uint8ToString(uint8(credit))
	_, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, creditCommand, creditHex), false)
	return err
}

// DownloadLogFile requests a number of log records to be downloaded.
func (ub *UbloxBluetooth) DownloadLogFile(cr *ConnectionReply, startingIndex int, fn DownloadLogHandler) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}
	si := uint16ToString(uint16(startingIndex))

	d, errr := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, readEventLogCommand, si), true)
	if errr != nil {
		return errors.Wrap(errr, "readEventLogCommand error")
	}

	expected, errr := ProcessEventsReply(d)
	if errr != nil {
		return errors.Wrap(errr, "ProcessEventsReply error")
	}

	errr = ub.SendCredits(cr, DefaultCredit)
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
					e = ub.SendCredits(cr, halfwayPoint)
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
func (ub *UbloxBluetooth) ClearEventLog(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, clearEventLogCommand), true)
	if err != nil {
		return errors.Wrap(err, "ClearEventLog error")
	}
	return ProcessClearEventReply(d)
}

// AbortEventLogRead aborts the read
func (ub *UbloxBluetooth) AbortEventLogRead(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	_, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, abortCommand), false)
	return err
}

// ReadSlotCount get recorder slot count
func (ub *UbloxBluetooth) ReadSlotCount(cr *ConnectionReply) (*SlotCountReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readSlotCountCommand), true)
	if err != nil {
		return nil, errors.Wrap(err, "ReadSlotCount error")
	}
	return NewSlotCountReply(d)
}

// ReadSlotInfo get recorder's slot info for the provided slotNumber, returns a SlotInfoReply structure or an error
func (ub *UbloxBluetooth) ReadSlotInfo(cr *ConnectionReply, slotNumber int) (*SlotInfoReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	slot := uint16ToString(uint16(slotNumber))
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, readSlotInfoCommand, slot), true)
	if err != nil {
		return nil, err
	}
	return NewSlotInfoReply(d)
}

// ReadSlotData gets the data for the given slot and offset
func (ub *UbloxBluetooth) ReadSlotData(cr *ConnectionReply, slotNumber int, offset int, requiredBytes int) ([]byte, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	slot := uint16ToString(uint16(slotNumber))
	off := uint16ToString(uint16(offset))
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, readSlotDataCommand, slot+off), true)
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
