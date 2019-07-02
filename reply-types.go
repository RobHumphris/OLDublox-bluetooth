package ubloxbluetooth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DiscoveryReply BLE discovery structure
type DiscoveryReply struct {
	BluetoothAddress string
	Rssi             int
	DeviceName       string
	DataType         int
	Data             string
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

// RS232SettingsReply serial port settings structure
type RS232SettingsReply struct {
	BaudRate           int
	FlowControl        int
	DataBits           int
	StopBits           int
	Parity             int
	ChangeAfterConfirm int
}

func NewRS232SettingsReply(d string) (*RS232SettingsReply, error) {
	var err error
	rsRply := RS232SettingsReply{}

	tokens := strings.Split(d, ",")
	if len(tokens) < 3 {
		return nil, fmt.Errorf("unknown response")
	}

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

// ConnectionReply connection data structure
type ConnectionReply struct {
	Handle           int
	Type             int
	BluetoothAddress string
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

// VersionReply VEH sensor version structure
type VersionReply struct {
	SoftwareVersion string
	HardwareVersion string
	ReleaseFlag     string
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

// ConfigReply sensor conf structure
type ConfigReply struct {
	AdvertisingInterval int
	SampleTime          int
	State               int
	AccelSettings       int
	SpareOne            int
	TemperatureOffset   int
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

// ConnectedPeer describes the Bluetooth peer's connection
type ConnectedPeer struct {
	PeerHandle int
	Type       int
	Profile    int
	MacAddress string
	FrameSize  int
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

// ACLConnected struct
type ACLConnected struct {
	ConnHandle int
	Type       int
	MacAddress string
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

// RecorderInfoReply structure
type RecorderInfoReply struct {
	SequenceNo int
	Count      int
	SlotUsage  int
	PoolUsage  int
}

// NewRecorderInfoReply takes the passed string and parses into a RecorderInfoReply
func NewRecorderInfoReply(s string) *RecorderInfoReply {
	return &RecorderInfoReply{
		SequenceNo: stringToInt(s[4:12]),
		Count:      stringToInt(s[12:16]),
		SlotUsage:  stringToInt(s[16:20]),
		PoolUsage:  stringToInt(s[20:24]),
	}
}

type RecorderMetaDataReply struct {
	Length int
	Crc    uint16
	Valid  bool
}

func NewRecorderMetaDataReply(s string) *RecorderMetaDataReply {
	valid := stringToInt(s[16:18])
	return &RecorderMetaDataReply{
		Length: stringToInt(s[4:12]),
		Crc:    uint16(stringToInt(s[12:16])),
		Valid:  valid > 0,
	}
}

// NewDownloadCRC returns an integer
func NewDownloadCRC(s string) int {
	return stringToInt(s[4:8])
}
