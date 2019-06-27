package ubloxbluetooth

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const unlockReply = "00"
const unlockSuccess = "0000"
const statusOk = "00"
const statusPending = "01"
const versionReply = "01"
const infoReply = "02"
const readConfigReply = "03"
const writeConfigReply = "04"
const readNameReply = "05"
const writeNameReply = "06"
const readEventLogReply = "07"
const clearEventLogReply = "08"
const readSlotCountReply = "0E"
const readSlotInfoReply = "0F"
const readSlotDataReply = "10"
const eraseSlotDataReply = "12"

const readRecorderInfoReply = "20"
const readRecorderReply = "21"
const queryRecorderMetaDataReply = "22"
const readRecorderDataReply = "23"

var readSlotDataReplyBytes = []byte(readSlotDataReply)
var readEventLogReplyBytes = []byte(readEventLogReply)

func isIndicationResponseValid(sa []string) bool {
	return sa[0] == "0" && sa[1] == "13"
}

func isNotificationResponseValid(nr [][]byte) bool {
	return nr[0][0] == 48 && nr[1][0] == 49 && nr[1][1] == 54
}

func splitOutResponse(d []byte, command string) (string, error) {
	b := bytes.Split(d, gattIndicationResponse)
	if len(b) < 2 {
		return "", fmt.Errorf("incorrect response")
	}
	tokens := strings.Split(string(b[1]), ",")
	if len(tokens) < 3 {
		return "", fmt.Errorf("unknown response")
	}
	if isIndicationResponseValid(tokens) {
		status := tokens[2][2:4]
		if tokens[2][0:2] == command && (status == statusOk || status == statusPending) {
			return tokens[2], nil
		}
	}
	return "", fmt.Errorf("invalid response")
}

func splitOutNotification(d []byte, command string) ([]byte, error) {
	b := bytes.Split(d, gattNotificationResponse)
	if len(b) < 2 {
		return nil, fmt.Errorf("incorrect response")
	}
	tokens := bytes.Split(b[1], comma)
	if len(tokens) < 3 {
		return nil, fmt.Errorf("unknown response")
	}
	if isNotificationResponseValid(tokens) {
		return tokens[2], nil
	}
	return nil, fmt.Errorf("invalid response")
}

func stringToHexString(s string) string {
	b, _ := hex.DecodeString(s)
	return string(b)
}

func stringToInt(s string) int {
	b, _ := hex.DecodeString(s)
	switch len(b) {
	case 1:
		return int(b[0])
	case 2:
		return int(binary.LittleEndian.Uint16(b))
	case 4:
		return int(binary.LittleEndian.Uint32(b))
	}
	return 0
}

func stringToFloat32(s string) float32 {
	b, _ := hex.DecodeString(s)
	intVal := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(intVal)
}

func uint8ToString(i uint8) string {
	b := make([]byte, 1)
	b[0] = i
	return hex.EncodeToString(b)
}

func uint16ToString(i uint16) string {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return hex.EncodeToString(b)
}

func uint32ToString(i uint32) string {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return hex.EncodeToString(b)
}

// NewDiscoveryReply takes the string and converts it to a DiscoveryReply
func NewDiscoveryReply(d string) (*DiscoveryReply, error) {
	t := strings.Split(d, ",")
	if len(t) < 5 {
		return nil, fmt.Errorf("[NewDiscoveryReply] Not enough tokens in string")
	}

	rssi, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrap(err, "[NewDiscoveryReply] error extracting RSSI")
	}

	dataType, err := strconv.Atoi(t[3])
	if err != nil {
		return nil, errors.Wrap(err, "[NewDiscoveryReply] error extracting DataType")
	}

	return &DiscoveryReply{
		BluetoothAddress: t[0],
		Rssi:             rssi,
		DeviceName:       t[2],
		DataType:         dataType,
		Data:             t[4],
	}, nil
}

// ProcessConnectDeviceSPSReply takes the passed string and attempts to parse it for its peerHandle
func ProcessConnectDeviceSPSReply(d string) (int, error) {
	b := strings.Split(d, connectPeerResponseString)
	if len(b) < 2 {
		return -1, fmt.Errorf("[ProcessConnectDeviceSPSReply] could not parse response (%s)", d)
	}

	handle, err := strconv.Atoi(b[1])
	if err != nil {
		return -1, errors.Wrapf(err, "[ProcessConnectDeviceSPSReply] error extracting Handle value (%s)", d)
	}
	return handle, nil
}

// ProcessPeerDisconnectedReply takes the passed string, pulls out the handle and checks that it matches the peerHandle
func ProcessPeerDisconnectedReply(peerHandle int, d string) error {
	b := strings.Split(d, disconnectPeerResponseString)
	if len(b) < 2 {
		return fmt.Errorf("[ProcessPeerDisconnectedReply] could not parse response (%s)", d)
	}

	handle, err := strconv.Atoi(b[1])
	if err != nil {
		return errors.Wrapf(err, "[ProcessPeerDisconnectedReply] error extracting Handle value (%s)", d)
	}

	if handle != peerHandle {
		return fmt.Errorf("[ProcessPeerDisconnectedReply] returned handle does not match required (wanted: %d got %d)", peerHandle, handle)
	}
	return nil
}

// NewConnectedPeerReply parses the passed string to assemble a new ConnectedPeer instance
func NewConnectedPeerReply(d string) (*ConnectedPeer, error) {
	b := strings.Split(d, peerConnectedResponseString)
	if len(b) < 2 {
		return nil, fmt.Errorf("[NewConnectedPeerReply] could not connect to device (%v)", b)
	}
	t := strings.Split(b[1], ",")
	if len(t) < 4 {
		return nil, fmt.Errorf("[NewConnectedPeerReply] could not connect to device (%v)", b)
	}

	handle, err := strconv.Atoi(t[0])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectedPeerReply] error extracting Handle value (%v)", b)
	}

	typ, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectedPeerReply] error extracting Type value (%v)", b)
	}

	profile, err := strconv.Atoi(t[2])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectedPeerReply] error extracting Profile value (%v)", b)
	}

	frameSize, err := strconv.Atoi(t[4])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectedPeerReply] error extracting FrameSize value (%v)", b)
	}

	return &ConnectedPeer{
		PeerHandle: handle,
		Type:       typ,
		Profile:    profile,
		MacAddress: t[3],
		FrameSize:  frameSize,
	}, nil
}

// NewACLConnectedReply parses the string for elements to extract and create a new ACLConnected instance
func NewACLConnectedReply(d string) (*ACLConnected, error) {
	b := strings.Split(d, aclConnectionRemoteDeviceResponseString)
	if len(b) < 2 {
		return nil, fmt.Errorf("[NewACLConnectedReply] could not connect to device (%v)", b)
	}
	t := strings.Split(b[1], ",")
	if len(t) < 3 {
		return nil, fmt.Errorf("[NewACLConnectedReply] could not connect to device (%v)", b)
	}

	connHandle, err := strconv.Atoi(t[0])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewACLConnectedReply] error extracting ConnHandle value (%v)", b)
	}

	typ, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectionReply] error extracting Type value (%v)", b)
	}
	return &ACLConnected{
		ConnHandle: connHandle,
		Type:       typ,
		MacAddress: t[2],
	}, nil
}

// NewConnectionReply takes the passed string and parses it into a Connection reply
func NewConnectionReply(d string) (*ConnectionReply, error) {
	b := strings.Split(d, connectResponse)
	if len(b) < 2 {
		return nil, fmt.Errorf("[NewConnectionReply] could not connect to device (%v)", b)
	}
	t := strings.Split(b[1], ",")
	if len(t) < 3 {
		return nil, fmt.Errorf("[NewConnectionReply] could not connect to device (%v)", b)
	}

	handle, err := strconv.Atoi(t[0])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectionReply] error extracting Handle value (%v)", b)
	}

	typ, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrapf(err, "[NewConnectionReply] error extracting Type value (%v)", b)
	}

	return &ConnectionReply{
		Handle:           handle,
		Type:             typ,
		BluetoothAddress: t[2],
	}, nil
}

// NewVersionReply returns a new VersionReply - or an error
func NewVersionReply(d []byte) (*VersionReply, error) {
	t, err := splitOutResponse(d, versionReply)
	if err != nil {
		return nil, err
	}

	return &VersionReply{
		SoftwareVersion: fmt.Sprintf("%d.%d", stringToInt(t[4:6]), stringToInt(t[6:8])),
		HardwareVersion: fmt.Sprintf("%d", stringToInt(t[8:10])),
		ReleaseFlag:     fmt.Sprintf("%d", stringToInt(t[10:12])),
	}, nil
}

// NewConfigReply returns a ConfigReply if the bytes are all present and correct, if not... an Error!
func NewConfigReply(d []byte) (*ConfigReply, error) {
	t, err := splitOutResponse(d, readConfigReply)
	if err != nil {
		return nil, err
	}
	return &ConfigReply{
		AdvertisingInterval: stringToInt(t[4:8]),
		SampleTime:          stringToInt(t[8:12]),
		State:               stringToInt(t[12:16]),
		AccelSettings:       stringToInt(t[16:20]),
		SpareOne:            stringToInt(t[20:24]),
		TemperatureOffset:   stringToInt(t[24:28]),
	}, nil
}

// ByteArray turns the `ConfigReply` to hex bytes
func (cr *ConfigReply) ByteArray() string {
	a := fmt.Sprintf("%s%s%s%s%s%s",
		uint16ToString(uint16(cr.AdvertisingInterval)),
		uint16ToString(uint16(cr.SampleTime)),
		uint16ToString(uint16(cr.State)),
		uint16ToString(uint16(cr.AccelSettings)),
		uint16ToString(uint16(cr.SpareOne)),
		uint16ToString(uint16(cr.TemperatureOffset)))
	return a
}

// NewNameReply returns the string value from the bytes in the response
func NewNameReply(d []byte) (string, error) {
	t, err := splitOutResponse(d, readNameReply)
	if err != nil {
		return "ERROR", err
	}

	name, err := hex.DecodeString(string(t[4:]))
	return string(name), err
}

// NewSlotCountReply returns a SlotCountReply
func NewSlotCountReply(d []byte) (*SlotCountReply, error) {
	t, err := splitOutResponse(d, readSlotCountReply)
	if err != nil {
		return nil, err
	}
	return &SlotCountReply{
		Count:    stringToInt(t[4:8]),
		rawCount: t[4:8],
	}, nil
}

/*
typedef struct
{
	// need to match recorder - whoops !!
	// not great here - so we will have to manually update !!!
	uint32_t time;
	uint16_t slot;
	uint16_t dwords;
	float    odr;  // should this be a float??
	uint16_t temp;
	uint16_t vbatt;
	uint16_t vin;

  } veh_slot_info_t;
*/ // NewSlotInfoReply returns a SlotInfoReply or error
func NewSlotInfoReply(d []byte) (*SlotInfoReply, error) {
	t, err := splitOutResponse(d, readSlotInfoReply)
	if err != nil {
		return nil, err
	}
	return &SlotInfoReply{
		Time:           stringToInt(t[4:12]),
		Slot:           stringToInt(t[12:16]),
		Bytes:          stringToInt(t[16:20]) * 4,
		SampleRate:     stringToFloat32(t[20:28]),
		Temperature:    stringToInt(t[28:32]),
		BatteryVoltage: stringToInt(t[32:36]),
		VoltageIn:      stringToInt(t[36:40]),
	}, nil
}

// ProcessUnlockReply returns true or false flag for unlock - or an error
func ProcessUnlockReply(d []byte) (bool, error) {
	t, err := splitOutResponse(d, unlockReply)
	if err != nil {
		return false, err
	}
	return (t == unlockSuccess), nil
}

// ProcessRS232SettingsReply processes the passed bytes for the RS232 settings
func ProcessRS232SettingsReply(d []byte) (*RS232SettingsReply, error) {
	var err error
	b := bytes.Split(d, rs232SettingsResponse)
	if len(b) < 2 {
		return nil, fmt.Errorf("incorrect response")
	}
	tokens := strings.Split(string(b[1]), ",")
	if len(tokens) < 3 {
		return nil, fmt.Errorf("unknown response")
	}

	rsRply := RS232SettingsReply{}

	rsRply.BaudRate, err = strconv.Atoi(tokens[0])
	if err != nil {
		return nil, errors.Wrap(err, "Baud conversion error")
	}

	rsRply.FlowControl, err = strconv.Atoi(tokens[1])
	if err != nil {
		return nil, errors.Wrap(err, "FlowControl conversion error")
	}

	rsRply.DataBits, err = strconv.Atoi(tokens[2])
	if err != nil {
		return nil, errors.Wrap(err, "DataBits conversion error")
	}

	rsRply.StopBits, err = strconv.Atoi(tokens[3])
	if err != nil {
		return nil, errors.Wrap(err, "StopBits conversion error")
	}

	rsRply.Parity, err = strconv.Atoi(tokens[4])
	if err != nil {
		return nil, errors.Wrap(err, "Parity conversion error")
	}

	return &rsRply, nil
}

// ErrUnexpectedResponse is a type of error that may not be catastrophic - just unexpected
var ErrUnexpectedResponse = fmt.Errorf("UnexpectedResponse")

// ProcessDiscoveryReply returns an array of DiscoveryReplys and a error
func ProcessDiscoveryReply(d []byte) (*DiscoveryReply, error) {
	b := bytes.Split(d, discoveryResponse)
	if len(b) < 1 {
		return nil, ErrUnexpectedResponse
	}
	return NewDiscoveryReply(string(b[1]))
}

// ProcessEventsReply returns the expected number of event notifications that we're about to receive.
func ProcessEventsReply(d []byte, reply string) (int, error) {
	t, err := splitOutResponse(d, reply)
	if err != nil {
		return -1, err
	}

	count := stringToInt(t[4:8])
	return count, nil
}

// ProcessClearEventReply checks the response and raises an error if things do not behave as they should.
func ProcessClearEventReply(d []byte) error {
	_, err := splitOutResponse(d, clearEventLogReply)
	return err
}

// ProcessEraseSlotDataReply check the passed response bytes.
func ProcessEraseSlotDataReply(d []byte) error {
	_, err := splitOutResponse(d, eraseSlotDataReply)
	return err
}

// ProcessSlotsReply returns a count of available slots.
func ProcessSlotsReply(d []byte) (int, error) {
	// +UUBTGI:0,13,10012603
	t, err := splitOutResponse(d, readSlotDataReply)
	if err != nil {
		return -1, err
	}
	count := stringToInt(t[4:8])
	return count, nil
}

// ProcessDisconnectReply checks the passed bytes for a correct disconnect.
func ProcessDisconnectReply(d []byte) (bool, error) {
	b := bytes.Split(d, disconnectResponse)
	if len(b) < 2 {
		return false, fmt.Errorf("disconnect error %q", d)
	}
	return b[1][0] == '0', nil
}

// ProcessRSSIReply - picks the data from the response in +UBTRSS:<rssi>
func ProcessRSSIReply(d []byte) (string, error) {
	b := bytes.Split(d, getRSSIResponse)
	if len(b) < 2 {
		return "??", fmt.Errorf("get RSSI error %q", d)
	}
	return string(b[1]), nil
}

// ProcessReadRecorderInfoReply - breaks down the response to extract RecorderInfo
func ProcessReadRecorderInfoReply(d []byte) (*RecorderInfoReply, error) {
	_, err := splitOutResponse(d, readRecorderInfoReply)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}

// ProcessReadRecorderReply - breaks down the response and
func ProcessReadRecorderReply(d []byte) (string, error) {
	t, err := splitOutResponse(d, readRecorderReply)
	if err != nil {
		return "nil", err
	}
	return t, fmt.Errorf("NOT IMPLEMENTED")
}

// ProcessQueryRecorderMetaDataReply - a
func ProcessQueryRecorderMetaDataReply(d []byte) (string, error) {
	t, err := splitOutResponse(d, queryRecorderMetaDataReply)
	if err != nil {
		return "nil", err
	}
	return t, fmt.Errorf("NOT IMPLEMENTED")
}

// ProcessReadRecorderDataReply - b
func ProcessReadRecorderDataReply(d []byte) (string, error) {
	t, err := splitOutResponse(d, readRecorderDataReply)
	if err != nil {
		return "nil", err
	}
	return t, fmt.Errorf("NOT IMPLEMENTED")
}
