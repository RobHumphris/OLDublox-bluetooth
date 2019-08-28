package ubloxbluetooth

import (
	"fmt"
	"time"

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

// RebootUblox reboots the Ublox chip
func (ub *UbloxBluetooth) RebootUblox() error {
	r := RebootCommand()
	err := ub.Write(r.Cmd)
	if err != nil {
		return err
	}
	//ub.currentMode = dataMode
	_, err = ub.WaitForResponse(r.Resp, false)
	modeSwitchDelay()
	return err
}

// GetDeviceRSSI gets the Recieved Signal Strength for the `address`
/*func (ub *UbloxBluetooth) GetDeviceRSSI(address string) (string, error) {
	d, err := ub.writeAndWait(GetRSSICommand(address), true)
	if err != nil {
		return "??", err
	}
	return ProcessRSSIReply(d)
}*/

// SetDTRBehavior configures the device to the correct DTR behaviour
func (ub *UbloxBluetooth) SetDTRBehavior() error {
	_, err := ub.writeAndWait(SetDTRBehaviorCommand(4), false)
	if err != nil {
		return errors.Wrap(err, "SetDTRBehaviorCommand error")
	}

	_, err = ub.writeAndWait(BLEStoreConfig(), false)
	if err != nil {
		return errors.Wrap(err, "BLEStoreConfig error")
	}
	return nil
}

// PeerList returns a list of connected peers.
func (ub *UbloxBluetooth) PeerList() error {
	d, err := ub.writeAndWait(PeerListCommand(), true)
	if err != nil {
		return err
	}

	fmt.Printf("RESULT: %s [%X]\n", d, d)

	return nil
}

// DiscoveryCommand issues the Discover command and calls the DiscoveryReplyHandler
// DiscoveryReplyHandler handles discovery replies
func (ub *UbloxBluetooth) DiscoveryCommand(timestamp int32, fn func(*DiscoveryReply, int32) error) error {
	dc := DiscoveryCommand()
	err := ub.Write(dc.Cmd)
	if err != nil {
		return err
	}

	return ub.HandleDiscovery(dc.Resp, func(d []byte) (bool, error) {
		dr, err := ProcessDiscoveryReply(d)
		if err == nil {
			err = fn(dr, timestamp)
		} else if err != ErrUnexpectedResponse {
			return false, err
		}
		return true, nil
	})
}

// ConnectToDevice attempts to connect to the device with the specified address.
func (ub *UbloxBluetooth) ConnectToDevice(address string, onConnect DeviceEvent, onDisconnect DeviceEvent) error {
	d, err := ub.writeAndWait(ConnectCommand(address), true)
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

// DisconnectFromDevice issues the disconnect command using the handle from the ConnectionReply
func (ub *UbloxBluetooth) DisconnectFromDevice() error {
	if ub.connectedDevice == nil {
		fmt.Printf("[DisconnectFromDevice] Device has been disconnected\n")
		return fmt.Errorf("ConnectionReply is nil")
	}
	ub.disconnectExpected = true

	d, err := ub.writeAndWait(DisconnectCommand(ub.connectedDevice.Handle), true)
	if err != nil {
		return err
	}
	fmt.Print("2")
	ok, err := ProcessDisconnectReply(d)
	if !ok {
		return fmt.Errorf("Incorrect disconnect reply %q", d)
	}
	fmt.Print("3\n")
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

// ReadCharacterisitic reads the connected device's BT Characteristics
func (ub *UbloxBluetooth) ReadCharacterisitic() ([]byte, error) {
	if ub.connectedDevice == nil {
		return nil, fmt.Errorf("ConnectionReply is nil")
	}
	d, err := ub.writeAndWait(ReadCharacterisiticCommand(ub.connectedDevice.Handle, commandValueHandle), true)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadCharacterisitic error")
	}
	return d, nil
}
