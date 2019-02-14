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

func stringToFloat32(s string) float32 {
	b, _ := hex.DecodeString(s)
	intVal := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(intVal)
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
		SoftwareVersion: stringToInt(t[4:8]),
		HardwareVersion: stringToInt(t[8:12]),
	}, nil
}

// NewInfoReply returns an InfoReply if the bytes are right, or an error if they're not
func NewInfoReply(d []byte) (*InfoReply, error) {
	t, err := splitOutResponse(d, infoReply)
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
	t, err := splitOutResponse(d, readConfigReply)
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
	fmt.Printf("Length: %d\n", len(d))
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

// ProcessEventsReply returns the expected number of event notifications that we're about to receive.
func ProcessEventsReply(d []byte) (int, error) {
	t, err := splitOutResponse(d, readEventLogReply)
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
