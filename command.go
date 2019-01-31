package ubloxbluetooth

import "fmt"

const at = "AT"

const powerOff = "+CPWROFF"

const discovery = "+UBTD"

const connect = "+UBTACLC"
const connectResponse = "+UUBTACLC:"

const disconnect = "+UBTACLD"
const disconnectResponse = "+UUBTACLD"

const writeCharacteristic = "+UBTGW"
const writeCharacteristicConfig = "+UBTGWC"

const errorMessage = "ERROR"
const okMessage = "OK"

const commandValueHandle = 13
const commandCCCDHandle = 14
const dataValueHandle = 16
const dataCCCDHandle = 17

var rebootResponse = []byte("+STARTUP")
var gattIndicationResponse = []byte("+UUBTGI:")
var discoveryResponse = []byte("+UBTD:")

var unlockCommand = []byte{0x00}
var versionCommand = []byte{0x01}
var infoCommand = []byte{0x02}
var readConfigCommand = []byte{0x03}
var writeConfigCommand = []byte{0x04}
var readNameCommand = []byte{0x05}
var writeNameCommand = []byte{0x06}
var readEventLogCommand = []byte{0x07}

var comma = []byte{0x2C}

func ATCommand() string {
	return at
}

func RebootCommand() string {
	return fmt.Sprintf("AT%s", powerOff)
}

func DiscoveryCommand() string {
	return fmt.Sprintf("AT%s", discovery)
}

func ConnectCommand(address string) string {
	return fmt.Sprintf("AT%s=%s", connect, address)
}

func DisconnectCommand(handle int) string {
	return fmt.Sprintf("AT%s=%d", disconnect, handle)
}

func WriteCharacteristicConfigurationCommand(connHandle int, descHandle int, config int) string {
	return fmt.Sprintf("AT%s=%d,%d,%d", writeCharacteristicConfig, connHandle, descHandle, config)
}

func WriteCharacteristicCommand(connHandle int, valueHandle int, data []byte) string {
	return fmt.Sprintf("AT%s=%d,%d,%x", writeCharacteristic, connHandle, valueHandle, data)
}
