package ubloxbluetooth

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

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

// EchoOff requests that the ublox device is a little less noisy
func (ub *UbloxBluetooth) EchoOff() error {
	_, err := ub.writeAndWait(EchoOffCommand(), false)
	return err
}

// GetRS232Settings allows us to see how the Ublox comms are configured
func (ub *UbloxBluetooth) GetRS232Settings() error {
	b, err := ub.writeAndWait(RS232SettingsCommand(""), true)
	fmt.Printf("Fishbin! %s\n", b)
	return err
}

// ConfigureUblox setups the ublox module
func (ub *UbloxBluetooth) ConfigureUblox() error {
	_, err := ub.writeAndWait(BLERole(bleCentral), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEConfig(minConnectionInterval, 6), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEConfig(maxConnectionInterval, 6), false)
	if err != nil {
		return err
	}

	_, err = ub.writeAndWait(BLEStoreConfig(), false)
	return err
}

// RebootUblox reboots the Ublox chip
func (ub *UbloxBluetooth) RebootUblox() error {
	_, err := ub.writeAndWait(RebootCommand(), true)
	if err != nil {
		return err
	}
	return nil
}

// DiscoveryCommand issues the Discover command and builds a list of new devices
func (ub *UbloxBluetooth) DiscoveryCommand() ([]DiscoveryReply, error) {
	d, err := ub.writeAndWait(DiscoveryCommand(), true)
	if err != nil {
		return nil, err
	}
	return ProcessDiscoveryReply(d)
}

// ConnectToDevice attempts to connect to the device with the specified address.
func (ub *UbloxBluetooth) ConnectToDevice(addr string) (*ConnectionReply, error) {
	d, err := ub.writeAndWait(ConnectCommand(addr), true)
	if err != nil {
		return nil, err
	}
	return NewConnectionReply(string(d))
}

// DisconnectFromDevice issues the disconnect command using the handle from the ConnectionReply
func (ub *UbloxBluetooth) DisconnectFromDevice(cr *ConnectionReply) error {
	if cr == nil {
		return fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(DisconnectCommand(cr.Handle), true)
	if err != nil {
		return err
	}

	ok, err := ProcessDisconnectReply(d)
	if !ok {
		return fmt.Errorf("Incorrect disconnect reply %q", d)
	}
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

// UnlockDevice attempts to unlock the device with the password provided.
func (ub *UbloxBluetooth) UnlockDevice(cr *ConnectionReply, password []byte) (bool, error) {
	if cr == nil {
		return false, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, append(unlockCommand, password...)), true)
	if err != nil {
		return false, err
	}
	return ProcessUnlockReply(d)
}

// GetVersion request the connected device's version
func (ub *UbloxBluetooth) GetVersion(cr *ConnectionReply) (*VersionReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, versionCommand), true)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return NewInfoReply(d)
}

// ReadConfig requests the device's current config
func (ub *UbloxBluetooth) ReadConfig(cr *ConnectionReply) (*ConfigReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readConfigCommand), true)
	if err != nil {
		return nil, err
	}
	return NewConfigReply(d)
}

// DownloadLogFile requests a number of log records to be downloaded.
func (ub *UbloxBluetooth) DownloadLogFile(cr *ConnectionReply, startingIndex int) ([][]byte, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}
	si := uint16ToString(uint16(startingIndex))
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, readEventLogCommand, si), true)
	if err != nil {
		return nil, err
	}

	expected, err := ProcessEventsReply(d)
	if err != nil {
		return nil, err
	}

	data := make([][]byte, expected)
	i := 0
	received, err := ub.HandleDataDownload(expected, func(d []byte) bool {
		if bytes.HasPrefix(d, gattNotificationResponse) {
			d, e := splitOutNotification(d, "07")
			if e != nil {
				err = errors.Wrapf(err, e.Error())
			} else {
				dt, e := hex.DecodeString(string(d[:]))
				if e != nil {
					err = errors.Wrapf(err, e.Error())
				} else {
					data[i] = dt
					i++
				}
			}
		} else if bytes.HasPrefix(d, gattIndicationResponse) {
			_, e := splitOutResponse(d, "07")
			err = errors.Wrapf(err, "notification error %v ", e)
			return false
		}
		return true
	})

	if received != expected {
		err = errors.Wrap(err, fmt.Sprintf("expected %d received %d\n", expected, received))
	}
	return data, err
}

// ReadSlotCount get recorder slot count
func (ub *UbloxBluetooth) ReadSlotCount(cr *ConnectionReply) (*SlotCountReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readSlotCountCommand), true)
	if err != nil {
		return nil, err
	}
	return NewSlotCountReply(d)
}

// ReadSlotInfo get recorder's slot info for the provided slotNumber, returns a SlotInfoReply structure or an error
func (ub *UbloxBluetooth) ReadSlotInfo(cr *ConnectionReply, slotNumber int) (*SlotInfoReply, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	d, err := ub.writeAndWait(WriteCharacteristicCommand(cr.Handle, commandValueHandle, readSlotInfoCommand), true)
	if err != nil {
		return nil, err
	}
	return NewSlotInfoReply(d)
}

// ReadSlotData gets the data for the given slot and offset
func (ub *UbloxBluetooth) ReadSlotData(cr *ConnectionReply, slotNumber int, offset int) ([]byte, error) {
	if cr == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}

	slot := uint16ToString(uint16(slotNumber))
	off := uint16ToString(uint16(offset))
	hex := slot + off
	d, err := ub.writeAndWait(WriteCharacteristicHexCommand(cr.Handle, commandValueHandle, readSlotDataCommand, hex), true)
	if err != nil {
		return nil, err
	}

	expected, err := ProcessSlotsReply(d)
	fmt.Printf("%q\nExpected %d\n", d, expected)
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}
