package ubloxbluetooth

import "fmt"

// CmdResp holds the AT CoMmanD and the expected RESPonse
type CmdResp struct {
	Cmd  string
	Resp string
}

// ATCommand - a simple AT message
func ATCommand() CmdResp {
	return CmdResp{
		Cmd:  at,
		Resp: empty,
	}
}

// EchoOffCommand ... commands the echo to turn off...
func EchoOffCommand() CmdResp {
	return CmdResp{
		Cmd:  echoOff,
		Resp: empty,
	}
}

// RS232SettingsCommand gets or sets the ublox serial port settings
func RS232SettingsCommand(cmd string) CmdResp {
	if cmd == "" {
		cmd = fmt.Sprintf("AT%s?", rs232Settings)
	} else {
		cmd = fmt.Sprintf("AT%s=%s", rs232Settings, cmd)
	}
	return CmdResp{
		Cmd:  cmd,
		Resp: rs232SettingsResponseString,
	}
}

// WatchdogCommand sets 't' typem and value
// where t is 1 value is milliseconds
// where t is 2 value is 0 or 1
func WatchdogCommand(t int, v int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d", watchdogSettings, t, v),
		Resp: empty,
	}
}

// FactoryResetCommand sets the ublox device to its Factory settings
func FactoryResetCommand() CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s", factoryReset),
		Resp: empty,
	}
}

// ModuleStartCommand sets the ublox device's Start mode
func ModuleStartCommand(mode StartMode) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d", moduleStartMode, mode),
		Resp: moduleStartModeResponseString,
	}
}

// RebootCommand - demands a reboot
func RebootCommand() CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s", powerOff),
		Resp: rebootResponseString,
	}
}

// GetRSSICommand - Returns the current Received signal strength for the device with the specified `address`
func GetRSSICommand(address string) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%s", getRSSI, address),
		Resp: getRSSIResponseString,
	}
}

// PeerListCommand - queries the connected Ublox device for all connected peers
func PeerListCommand() CmdResp {
	return CmdResp{
		Cmd:  peerList,
		Resp: peerListResponseString,
	}
}

// DiscoveryCommand - commands that all is discovered
func DiscoveryCommand(scanMS int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=4,1,%d", discovery, scanMS),
		Resp: discoveryResponseString,
	}
}

// BLERole - for setting the role with one of the following constants:
// bleDisabled  0
// bleCentral 1
// blePeripheral 2
// bleSimultaneous  3
func BLERole(role int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d", bleRole, role),
		Resp: empty,
	}
}

// SetDTRBehaviorCommand is called with one of the following values:
// 0 ignore line
// 1 Assert->Deassert = Enters command mode
// 2 Assert->Deassert = Orderly disconnect of all radio links
// 3 Assert->Deassert = UART deactivated. Reactivate with Deassert->Assert
//   or BT connection
// 4 Assert->Deassert = Module Shut off. Deassert->Assert start again.
func SetDTRBehaviorCommand(value int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT&D%d", value),
		Resp: empty,
	}
}

// BLEConfig sets the Bluetooth LE config (see: 6.26.3 Defined values) from:
// https://www.u-blox.com/sites/default/files/u-blox-SHO_ATCommands_%28UBX-14044127%29.pdf
func BLEConfig(param int, val int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d", bleConfiguration, param, val),
		Resp: empty,
	}
}

// BLEStoreConfig follows the BLEConfig commands, these only take effect after
// the RebootCommand() is issued.
func BLEStoreConfig() CmdResp {
	return CmdResp{
		Cmd:  storeConfig,
		Resp: empty,
	}
}

// ConnectCommand Constructs the command to connect to a device
func ConnectCommand(address string) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%s", connect, address),
		Resp: connectResponse,
	}
}

// DisconnectCommand Constructs the command to disconnect to a device
func DisconnectCommand(handle int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d", disconnect, handle),
		Resp: disconnectResponseString,
	}
}

// WriteCharacteristicConfigurationCommand constructs the command to set device characteristics
func WriteCharacteristicConfigurationCommand(connHandle int, descHandle int, config int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%d", writeCharacteristicConfig, connHandle, descHandle, config),
		Resp: empty,
	}
}

// ReadCharacterisiticCommand constructs the command to read the device's characteristics
func ReadCharacterisiticCommand(connHandle int, valueHandle int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d", readCharacterisitic, connHandle, valueHandle),
		Resp: empty,
	}
}

type characteristicCommand struct {
	connectionHandle int
	valueHandle      int
	data             []byte
}

func writeCharacteristicCommand(c characteristicCommand) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%x", writeCharacteristic, c.connectionHandle, c.valueHandle, c.data),
		Resp: gattIndicationResponseString,
	}
}

type characteristicHexCommand struct {
	*characteristicCommand
	hex string
}

func writeCharacteristicHexCommand(c characteristicHexCommand) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%x%s", writeCharacteristic, c.connectionHandle, c.valueHandle, c.data, c.hex),
		Resp: gattIndicationResponseString,
	}
}

// ConnectPeerCommand creates the command
func ConnectPeerCommand(url string) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%s", connectPeer, url),
		Resp: connectPeerResponseString,
	}
}

// DisconnectPeerCommand creates the command
func DisconnectPeerCommand(peerHandle int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d", disconnectPeer, peerHandle),
		Resp: disconnectPeerResponseString,
	}
}

// EnterDataModeCommand creates the command
func EnterDataModeCommand() CmdResp {
	return CmdResp{
		Cmd:  enterDataMode,
		Resp: empty,
	}
}

// EnterExtendedDataModeCommand creates the command
func EnterExtendedDataModeCommand() CmdResp {
	return CmdResp{
		Cmd:  enterExtendedDataMode,
		Resp: empty,
	}
}

// IssueEscapeSequence creates the command
func IssueEscapeSequence() CmdResp {
	return CmdResp{
		Cmd:  escapeSequence,
		Resp: empty,
	}
}

// GetSerialCommand creates the command
func GetSerialCommand() CmdResp {
	return CmdResp{
		Cmd:  at + getSerialCmd,
		Resp: quotedStringResponse,
	}
}
