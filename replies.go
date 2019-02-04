package ubloxbluetooth

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type DiscoveryReply struct {
	BluetoothAddress string
	Rssi             int
	DeviceName       string
	DataType         int
	Data             string
}

type ConnectionReply struct {
	Handle           int
	Type             int
	BluetoothAddress string
}

type VersionReply struct {
	SoftwareVersion int
	HardwareVersion int
}

type InfoReply struct {
	CurrentTime           int
	CurrentSequenceNumber int
	RecordsCount          int
}

type ConfigReply struct {
	AdvertisingInterval int
	SampleTime          int
	State               int
	AccelSettings       int
	SpareOne            int
	TemperatureOffset   int
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

// NewConnectionReply takes the passed string and parses it into a Connection reply
func NewConnectionReply(d string) (*ConnectionReply, error) {
	b := strings.Split(d, connectResponse)
	if len(b) < 2 {
		return nil, fmt.Errorf("[NewConnectionReply] could not connect to device")
	}
	t := strings.Split(b[1], ",")
	if len(t) < 3 {
		return nil, fmt.Errorf("[NewConnectionReply] could not connect to device")
	}

	handle, err := strconv.Atoi(t[0])
	if err != nil {
		return nil, errors.Wrap(err, "[NewConnectionReply] error extracting Handle value")
	}

	typ, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrap(err, "[NewConnectionReply] error extracting Type value")
	}

	return &ConnectionReply{
		Handle:           handle,
		Type:             typ,
		BluetoothAddress: t[2],
	}, nil
}

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
		if tokens[2][0:2] == command && (status == "00" || status == "01") {
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

func stringToInt(s string) int {
	b, _ := hex.DecodeString(s)
	switch len(b) {
	case 2:
		return int(binary.LittleEndian.Uint16(b))
	case 4:
		return int(binary.LittleEndian.Uint32(b))
	}
	return 0
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

// ProcessUnlockReply returns true or false flag for unlock - or an error
func ProcessUnlockReply(d []byte) (bool, error) {
	t, err := splitOutResponse(d, "00")
	if err != nil {
		return false, err
	}
	return (t == "0000"), nil
}

// ProcessDiscoveryReply returns an array of DiscoveryReplys and a error
func ProcessDiscoveryReply(d []byte) ([]DiscoveryReply, error) {
	var err error
	discovered := []DiscoveryReply{}
	b := bytes.Split(d, discoveryResponse)
	for i := 0; i < len(b); i++ {
		if len(b[i]) > 0 {
			d, e := NewDiscoveryReply(string(b[i]))
			if e != nil {
				err = errors.Wrapf(err, "NewDiscoveryReply error: %v\n", e)
			} else {
				discovered = append(discovered, *d)
			}
		}
	}
	return discovered, err
}

// NewVersionReply returns a new VersionReply - or an error
func NewVersionReply(d []byte) (*VersionReply, error) {
	t, err := splitOutResponse(d, "01")
	if err != nil {
		return nil, err
	}

	return &VersionReply{
		SoftwareVersion: stringToInt(t[4:8]),
		HardwareVersion: stringToInt(t[8:12]),
	}, nil
}

// NewInfoReply returns an InfoReply if the bytes are right, or an error if they're not
func NewInfoReply(d []byte) (*InfoReply, error) {
	t, err := splitOutResponse(d, "02")
	if err != nil {
		return nil, err
	}

	return &InfoReply{
		CurrentTime:           stringToInt(t[4:12]),
		CurrentSequenceNumber: stringToInt(t[12:16]),
		RecordsCount:          stringToInt(t[16:20]),
	}, nil
}

// NewConfigReply returns a ConfigReply if the bytes are all present and correct, if not... an Error!
func NewConfigReply(d []byte) (*ConfigReply, error) {
	t, err := splitOutResponse(d, "03")
	if err != nil {
		return nil, err
	}
	return &ConfigReply{
		AdvertisingInterval: stringToInt(t[4:8]),
		SampleTime:          stringToInt(t[8:12]),
		State:               stringToInt(t[12:16]),
		AccelSettings:       stringToInt(t[16:18]),
		SpareOne:            stringToInt(t[18:20]),
		TemperatureOffset:   stringToInt(t[20:22]),
	}, nil
}

// ProcessEventsReply returns the expected number of event notifications that we're about to receive.
func ProcessEventsReply(d []byte) (int, error) {
	t, err := splitOutResponse(d, "07")
	if err != nil {
		return -1, err
	}

	count := stringToInt(t[4:8])
	return count, nil
}
