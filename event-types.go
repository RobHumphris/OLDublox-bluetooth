package ubloxbluetooth

import (
	"encoding/binary"
	"fmt"
	"math"
)

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
	// VehEventError
}

// NewRecorderEvent returns new RecorderEvent
func NewRecorderEvent(b []byte) (*VehEvent, error) {
	fn, ok := handlers[b[8]]
	if ok {
		return fn(b), nil
	}
	return nil, fmt.Errorf("Unhandled Event type: %02X", b[8])
}

func newVehEvent(b []byte) (VehEvent, int) {
	length := 10 + int(b[9])
	return VehEvent{
		DataFlag:  b[length] > 0x00,
		Sequence:  binary.LittleEndian.Uint32(b[0:4]),
		Timestamp: binary.LittleEndian.Uint32(b[4:8]),
		EventType: int(b[8]),
	}, length
}

func newBootEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.BootEvent = &VehBootEvent{
		binary.LittleEndian.Uint32(b[10:14]),
		fmt.Sprintf("%d.%d", b[14], b[15]),
		uint32(b[16]),
		uint32(b[17]),
	}
	return &eb
}

func newSensorEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.SensorEvent = &VehSensorEvent{
		float32(binary.LittleEndian.Uint16(b[10:12]) / 4),
		float32(binary.LittleEndian.Uint16(b[12:14]) / 1000),
		make([]byte, l-14),
	}
	copy(eb.SensorEvent.Other, b[13:l-1])
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
		Battery:         float32FromBytes(b[10:14]),
		Temperature:     float32FromBytes(b[14:18]),
		Humidity:        float32FromBytes(b[18:22]),
		SoftTemperature: float32FromBytes(b[22:26]),
	}
	return &eb
}

func newVibrationEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.VibrationEvent = &VehVibrationEvent{
		Battery:     float32FromBytes(b[10:14]),
		Temperature: float32FromBytes(b[14:18]),
		Odr:         float32FromBytes(b[18:22]),
		Gain:        float32FromBytes(b[22:26]),
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
