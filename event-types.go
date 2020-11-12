package ubloxbluetooth

import (
	"encoding/binary"
	"fmt"
	"math"
)

// VEH Sensor event types
const (
	VehEventBoot         = 0x00
	VehEventSensor       = 0x02
	VehEventMessage      = 0x0D
	VehEventDummy        = 0x16
	VehEventTemperature  = 0x64
	VehEventVibration    = 0x65
	VehEventConnected    = 0xC0
	VehEventDisconnected = 0xD0
	VehEventSystemOff    = 0x00
	VehEventError        = 0xFF
)

type field struct {
	offset int
	size   int
}

func (f *field) Start() int {
	return f.offset
}

func (f *field) End() int {
	return f.offset + f.size
}

const (
	sizeofFloat32 = 4
	sizeofUint32  = 4
	sizeofUint16  = 2
	sizeofUint8   = 1
	variableSize  = 0
)

var (
	/*
		VEH_EVT_HDR
		Offset		Parameter		Type			Description
		0			seqno			uint32_t		Sequence number
		4			timestamp		uint32_t		Timestamp in seconds
		8			tag				uint8_t			Event type
		9			length			uint8_t			Length of payload (bytes)
		10			Payload			uint8_t[]		Event payload
		10+length	flag			uint8_t			Flag to indicate presence of data
	*/
	ehSeqNoOffset     field = field{0, sizeofUint32}
	ehTimestampOffset field = field{4, sizeofUint32}
	ehEventTypeOffset field = field{8, sizeofUint8}
	ehLengthOffset    field = field{9, sizeofUint8}
	ehPayloadOffset   field = field{10, variableSize}

	/*
		VEH_EVT_BOOT
		Offset		Parameter		Type			Description
		10			reason			uint32_t		CPU reset register contents
		14			version			uint8[4]		Version
		18			flag			uint8_t			0 = no data
	*/
	ebReasonOffset  field = field{10, sizeofUint32}
	ebVersionOffset field = field{14, sizeofUint32}
	ebFlagOffset    field = field{18, sizeofUint8}

	/*
		VEH_EVT_SENSOR - No documentation exists for this!
		Offset		Parameter		Type			Description
		10			temperature		uint16_t		Temperature reading
		12			battery			uint16_t		Battery reading in mV
		14			other			uint8[length-14]?
		?			flag			uint8_t			0 = no data
	*/
	esTemperatureOffset field = field{10, sizeofUint16}
	esBatteryOffset     field = field{12, sizeofUint16}
	esOtherOffset       field = field{14, variableSize}

	/*
		VEH_EVT_APP_TEMPERATURE
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			humidity		float32_t		Relative humidity (%)
		22			sd_temp			float32_t		Soft Device temperature (C)
		26			flag			uint8_t			0 = no data
	*/
	etBatteryOffset     field = field{10, sizeofFloat32}
	etTemperatureOffset field = field{14, sizeofFloat32}
	etHumidityOffset    field = field{18, sizeofFloat32}
	etSdTempOffset      field = field{22, sizeofFloat32}
	etFlagOffset        field = field{26, sizeofUint8}

	/*
		VEH_EVT_APP_VIBRATION
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			odr				float32_t		Estimated sample rate freq (Hz)
		22			gain			float32_t		Gain setting
		26			other			float32_t		Spare
		30			flags			float32_t		Spare
		34			flag			uint8_t			0 = no data, 1 = data
	*/
	evBatteryOffset     field = field{10, sizeofFloat32}
	evTemperatureOffset field = field{14, sizeofFloat32}
	evOdrOffset         field = field{18, sizeofFloat32}
	evGainOffset        field = field{22, sizeofFloat32}
	evOtherOffset       field = field{26, sizeofFloat32}
	evFlagsOffset       field = field{30, sizeofFloat32}
	evFlagOffset        field = field{34, sizeofUint8}
)

type handler func([]byte) *VehEvent
type handlersMap map[byte]handler

var handlers handlersMap

// VehEvent is the base interface
type VehEvent struct {
	DataFlag          bool
	Sequence          uint32
	Timestamp         uint32
	EventType         int
	BootEvent         *VehBootEvent
	SensorEvent       *VehSensorEvent
	ConnectedEvent    *VehConnectedEvent
	DisconnectedEvent *VehDisconnectedEvent
	TemperatureEvent  *VehTemperatureEvent
	VibrationEvent    *VehVibrationEvent
}

// VehBootEvent structure
type VehBootEvent struct {
	Reason          uint32
	SoftwareVersion string
	HardwareVersion uint32
	BuildNumber     uint32
}

// VehSensorEvent structure
type VehSensorEvent struct {
	Temperature       float32
	BatteryMilliVolts float32
	Other             []byte
}

// VehConnectedEvent structure
type VehConnectedEvent struct {
	Mac string
}

// VehDisconnectedEvent structure
type VehDisconnectedEvent struct {
	Mac string
}

// VehTemperatureEvent structure
type VehTemperatureEvent struct {
	Battery         float32
	Temperature     float32
	Humidity        float32
	SoftTemperature float32
}

// VehVibrationEvent structure
type VehVibrationEvent struct {
	Battery     float32
	Temperature float32
	Odr         float32
	Gain        float32
}

func init() {
	handlers = make(handlersMap)
	handlers[VehEventBoot] = newBootEvent
	handlers[VehEventSensor] = newSensorEvent
	// VehEventMessage
	// VehEventDummy
	handlers[VehEventTemperature] = newTemperatureEvent
	handlers[VehEventVibration] = newVibrationEvent
	handlers[VehEventConnected] = newConnectedEvent
	handlers[VehEventDisconnected] = newDisconnectedEvent
	// VehEventSystemOff
	handlers[VehEventError] = newErrorEvent
}

// NewRecorderEvent returns new RecorderEvent
func NewRecorderEvent(b []byte) (*VehEvent, error) {
	if len(b) < ehLengthOffset.Start() {
		return nil, ErrorTruncatedResponse
	}

	fn, ok := handlers[b[ehEventTypeOffset.Start()]]
	if ok {
		return fn(b), nil
	}
	return nil, fmt.Errorf("Unhandled Event type: %02X", b[ehEventTypeOffset.Start()])
}

func newErrorEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	return &eb
}

func newVehEvent(b []byte) (VehEvent, int) {
	length := 10 + int(b[ehLengthOffset.Start()])
	return VehEvent{
		DataFlag:  b[length] > 0x00,
		Sequence:  binary.LittleEndian.Uint32(b[ehSeqNoOffset.Start():ehSeqNoOffset.End()]),
		Timestamp: binary.LittleEndian.Uint32(b[ehTimestampOffset.Start():ehTimestampOffset.End()]),
		EventType: int(b[ehEventTypeOffset.Start()]),
	}, length
}

func newBootEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.BootEvent = &VehBootEvent{
		binary.LittleEndian.Uint32(b[ebReasonOffset.Start():ebReasonOffset.End()]),
		fmt.Sprintf("%d.%d", b[ebVersionOffset.Start()], b[ebVersionOffset.Start()+1]),
		uint32(b[ebVersionOffset.Start()+2]),
		uint32(b[ebVersionOffset.Start()+3]),
	}
	return &eb
}

func newSensorEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.SensorEvent = &VehSensorEvent{
		float32(binary.LittleEndian.Uint16(b[esTemperatureOffset.Start():esTemperatureOffset.End()]) / 4),
		float32(binary.LittleEndian.Uint16(b[esBatteryOffset.Start():esBatteryOffset.End()]) / 1000),
		make([]byte, l-14),
	}
	copy(eb.SensorEvent.Other, b[esOtherOffset.Start()-1:l-1])
	return &eb
}

func newConnectedEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.ConnectedEvent = &VehConnectedEvent{
		macString(b, l),
	}
	return &eb
}

func newDisconnectedEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.DisconnectedEvent = &VehDisconnectedEvent{
		macString(b, l),
	}
	return &eb
}

func newTemperatureEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.TemperatureEvent = &VehTemperatureEvent{
		Battery:         float32FromBytes(b[etBatteryOffset.Start():etBatteryOffset.End()]),
		Temperature:     float32FromBytes(b[etTemperatureOffset.Start():etTemperatureOffset.End()]),
		Humidity:        float32FromBytes(b[etHumidityOffset.Start():etHumidityOffset.End()]),
		SoftTemperature: float32FromBytes(b[etSdTempOffset.Start():etSdTempOffset.End()]),
	}
	return &eb
}

func newVibrationEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.VibrationEvent = &VehVibrationEvent{
		Battery:     float32FromBytes(b[evBatteryOffset.Start():evBatteryOffset.End()]),
		Temperature: float32FromBytes(b[evTemperatureOffset.Start():evTemperatureOffset.End()]),
		Odr:         float32FromBytes(b[evOdrOffset.Start():evOdrOffset.End()]),
		Gain:        float32FromBytes(b[evGainOffset.Start():evGainOffset.End()]),
	}
	return &eb
}

func macString(b []byte, l int) string {
	return fmt.Sprintf("%X:%X:%X:%X:%X:%X", b[l-1], b[l-2], b[l-3], b[l-4], b[l-5], b[l-6])
}

func float32FromBytes(b []byte) float32 {
	intVal := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(intVal)
}
