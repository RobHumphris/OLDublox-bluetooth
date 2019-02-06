package ubloxbluetooth

import "fmt"

const at = "AT"

const powerOff = "+CPWROFF"
const rebootResponse = "+STARTUP"

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

var gattIndicationResponse = []byte(gattIndicationResponseString)
var gattNotificationResponse = []byte("+UUBTGN:")

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

var comma = []byte{0x2C}

// ATCommand - a simple AT message
func ATCommand() string {
	return at
}

// RebootCommand - demands a reboot
func RebootCommand() string {
	return fmt.Sprintf("AT%s", powerOff)
}

func DiscoveryCommand() (string, string) {
	return fmt.Sprintf("AT%s", discovery), discoveryResponseString
}

func BLERole(role int) string {
	return fmt.Sprintf("AT%s=%d", bleRole, role)
}

func BLEConfig(param int, val int) string {
	return fmt.Sprintf("AT%s=%d,%d", bleConfiguration, param, val)
}

func BLEStoreConfig() string {
	return "AT&W"
}

func ConnectCommand(address string) (string, string) {
	return fmt.Sprintf("AT%s=%s", connect, address), connectResponse
}

func DisconnectCommand(handle int) (string, string) {
	return fmt.Sprintf("AT%s=%d", disconnect, handle), disconnectResponseString
}

func WriteCharacteristicConfigurationCommand(connHandle int, descHandle int, config int) string {
	return fmt.Sprintf("AT%s=%d,%d,%d", writeCharacteristicConfig, connHandle, descHandle, config)
}

func WriteCharacteristicCommand(connHandle int, valueHandle int, data []byte) (string, string) {
	return fmt.Sprintf("AT%s=%d,%d,%x", writeCharacteristic, connHandle, valueHandle, data), gattIndicationResponseString
}

func WriteCharacteristicHexCommand(connHandle int, valueHandle int, data []byte, hex string) (string, string) {
	return fmt.Sprintf("AT%s=%d,%d,%x%s", writeCharacteristic, connHandle, valueHandle, data, hex), gattIndicationResponseString
}
