package ubloxbluetooth

/**
* The commands and responses defined here are described in the u-blox
* document found here https://www.u-blox.com/sites/default/files/u-blox-SHO_ATCommands_%28UBX-14044127%29.pdf
 */
const empty = ""
const newline = "\r\n"
const at = "AT"
const rs232Settings = "+UMRS"
const rs232SettingsResponseString = "+UMRS:"

var rs232SettingsResponse = []byte(rs232SettingsResponseString)

const echoOff = "ATE0"
const storeConfig = "AT&W"

const powerOff = "+CPWROFF"
const rebootResponseString = "+STARTUP"
const factoryReset = "+UFACTORY"
const moduleStartMode = "+UMSM"
const moduleStartModeResponseString = "+UMSM:"

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

const readCharacterisitic = "+UBTGR"

const connectPeer = "+UDCP"
const connectPeerResponseString = "+UDCP:"
const peerConnectedResponseString = "+UUDPC:"
const aclConnectionRemoteDeviceResponseString = "+UUBTACLC:"

const disconnectPeer = "+UDCPC"
const disconnectPeerResponseString = "+UUDPD:"

const readEscapeCharacter = "S2"
const escapeSequence = "+++"
const enterDataMode = "ATO1"
const enterExtendedDataMode = "ATO2"

const errorMessage = "ERROR"
const okMessage = "OK"

const commandValueHandle = 13
const commandCCCDHandle = 14
const dataValueHandle = 16
const dataCCCDHandle = 17

const _SPSCharacteristic = "2456e1b9-26e2-8f83-e744-f34f01e9d701"
const _FifoCharacteristic = "2456e1b9-26e2-8f83-e744-f34f01e9d703"
const _CreditsCharacteristic = "2456e1b9-26e2-8f83-e744-f34f01e9d704"

const gattIndicationResponseString = "+UUBTGI:"

var ubloxBTReponseHeader = []byte("+UUBT")
var gattIndicationResponse = []byte(gattIndicationResponseString)
var gattNotificationResponse = []byte("+UUBTGN:")
var blePHYUpdateResponse = []byte("+UUBTLEPHYU:")
var peerConnectedResponse = []byte(peerConnectedResponseString)
var aclConnectionRemoteDeviceResponse = []byte(aclConnectionRemoteDeviceResponseString)

var tail = []byte{'\r', '\n'}
var separator = []byte(":")
var comma = []byte{0x2C}
