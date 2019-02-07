package ubloxbluetooth

import "fmt"

/**
* The commands and responses defined here are described in the u-blox
* document found here https://www.u-blox.com/sites/default/files/u-blox-SHO_ATCommands_%28UBX-14044127%29.pdf
 */
const empty = ""
const newline = "\r\n"
const at = "AT"
const rs232Settings = "+UMRS"
const echoOff = "ATE"
const storeConfig = "AT&W"

const powerOff = "+CPWROFF"
const rebootResponseString = "+STARTUP"

var rebootResponse = []byte(rebootResponseString)

const discovery = "+UBTD"
const discoveryResponseString = "+UBTD:"

var discoveryResponse = []byte(discoveryResponseString)

const bleRole = "+UBTLE"
const bleDisabled = 0
const bleCentral = 1
const blePeripheral = 2
const bleSimultaneous = 3

const bleConfiguration = "+UBTLECFG"
const minConnectionInterval = 4
const maxConnectionInterval = 5

const connect = "+UBTACLC"
const connectResponse = "+UUBTACLC:"

const disconnect = "+UBTACLD"
const disconnectResponseString = "+UUBTACLD:"

var disconnectResponse = []byte(disconnectResponseString)

const writeCharacteristic = "+UBTGW"
const writeCharacteristicResponseString = ""
const writeCharacteristicConfig = "+UBTGWC"

const errorMessage = "ERROR"
const okMessage = "OK"

const commandValueHandle = 13
const commandCCCDHandle = 14
const dataValueHandle = 16
const dataCCCDHandle = 17

const gattIndicationResponseString = "+UUBTGI:"

var ubloxBTReponseHeader = []byte("+UUBT")
var gattIndicationResponse = []byte(gattIndicationResponseString)
var gattNotificationResponse = []byte("+UUBTGN:")
var blePHYUpdateResponse = []byte("+UUBTLEPHYU:")

var unlockCommand = []byte{0x00}
var versionCommand = []byte{0x01}
var infoCommand = []byte{0x02}
var readConfigCommand = []byte{0x03}
var writeConfigCommand = []byte{0x04}
var readNameCommand = []byte{0x05}
var writeNameCommand = []byte{0x06}
var readEventLogCommand = []byte{0x07}
var readSlotCountCommand = []byte{0x0E}
var readSlotInfoCommand = []byte{0x0F}
var readSlotDataCommand = []byte{0x10}

var tail = []byte{'\r', '\n'}
var separator = []byte(":")
var comma = []byte{0x2C}

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
	r := CmdResp{}
	if cmd == "" {
		r.Cmd = fmt.Sprintf("AT%s?", rs232Settings)
	}
	r.Cmd = fmt.Sprintf("AT%s=%s", rs232Settings, cmd)
	r.Resp = empty
	return r
}

// RebootCommand - demands a reboot
func RebootCommand() CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s", powerOff),
		Resp: rebootResponseString,
	}
}

// DiscoveryCommand - commands that all is discovered
func DiscoveryCommand() CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s", discovery),
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

func ConnectCommand(address string) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%s", connect, address),
		Resp: connectResponse,
	}
}

func DisconnectCommand(handle int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d", disconnect, handle),
		Resp: disconnectResponseString,
	}
}

func WriteCharacteristicConfigurationCommand(connHandle int, descHandle int, config int) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%d", writeCharacteristicConfig, connHandle, descHandle, config),
		Resp: empty,
	}
}

func WriteCharacteristicCommand(connHandle int, valueHandle int, data []byte) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%x", writeCharacteristic, connHandle, valueHandle, data),
		Resp: gattIndicationResponseString,
	}
}

func WriteCharacteristicHexCommand(connHandle int, valueHandle int, data []byte, hex string) CmdResp {
	return CmdResp{
		Cmd:  fmt.Sprintf("AT%s=%d,%d,%x%s", writeCharacteristic, connHandle, valueHandle, data, hex),
		Resp: gattIndicationResponseString,
	}
}
