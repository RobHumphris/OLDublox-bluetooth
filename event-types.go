package ubloxbluetooth

import (
	"encoding/binary"
	"fmt"
)

const (
	vehEventBoot         = 0x00
	vehEventSensor       = 0x02
	vehEventMessage      = 0x0D
	vehEventDummy        = 0x16
	vehEventVibration    = 0x17
	vehEventConnected    = 0xC0
	vehEventDisconnected = 0xD0
	vehEventSystemOff    = 0x00
	vehEventError        = 0xFF
)

// Event is the base interface
type Event interface {
	DataFlag() bool
	Sequence() int
}

type handler func([]byte) Event
type handlersMap map[byte]handler

var handlers handlersMap

// RecorderEvents structure
type RecorderEvents struct {
	Events             []Event
	DataEventSequences []int
}

func init() {
	handlers = make(handlersMap)
	handlers[vehEventBoot] = newBootEvent
	handlers[vehEventSensor] = newSensorEvent
	// vehEventMessage
	// vehEventDummy
	handlers[vehEventVibration] = newVibrationEvent
	handlers[vehEventConnected] = newConnectedEvent
	handlers[vehEventDisconnected] = newDisconnectedEvent
	// vehEventSystemOff
	// vehEventError

}

// NewRecorderEvent returns new RecorderEvent
func NewRecorderEvent(b []byte) Event {
	fn, ok := handlers[b[8]]
	if ok {
		return fn(b)
	}
	fmt.Printf("UNHANDLED: %X\t[%X]\n", b[8], b)
	return nil
}

// EventBase structure
type EventBase struct {
	dataFlag  bool
	sequence  int
	Timestamp int
}

// DataFlag shows if this event has associated data
func (o *EventBase) DataFlag() bool {
	return o.dataFlag
}

// Sequence returns the event's sequence number
func (o *EventBase) Sequence() int {
	return o.sequence
}

func newEventBase(b []byte) (*EventBase, int) {
	length := 10 + int(b[9])
	return &EventBase{
		dataFlag:  b[length] > 0x00,
		sequence:  int(binary.LittleEndian.Uint32(b[0:4])),
		Timestamp: int(binary.LittleEndian.Uint32(b[4:8])),
	}, length
}

// BootEvent structure
type BootEvent struct {
	*EventBase
	Reason          int
	SoftwareVersion string
	HardwareVersion int
	BuildNumber     int
}

func newBootEvent(b []byte) Event {
	eb, _ := newEventBase(b)
	return &BootEvent{
		eb,
		int(binary.LittleEndian.Uint32(b[10:14])),
		fmt.Sprintf("%d.%d", b[14], b[15]),
		int(b[16]),
		int(b[17]),
	}
}

// SensorEvent structure
type SensorEvent struct {
	*EventBase
	Temperature       float32
	BatteryMilliVolts float32
	Other             []byte
}

func newSensorEvent(b []byte) Event {
	eb, l := newEventBase(b)
	o := &SensorEvent{
		eb,
		float32(binary.LittleEndian.Uint16(b[10:12]) / 4),
		float32(binary.LittleEndian.Uint16(b[12:14]) / 1000),
		make([]byte, l-14),
	}
	copy(o.Other, b[13:l-1])
	return o
}

// ConnectedEvent structure
type ConnectedEvent struct {
	*EventBase
	Mac string
}

func macString(b []byte, l int) string {
	return fmt.Sprintf("%X:%X:%X:%X:%X:%X", b[l-1], b[l-2], b[l-3], b[l-4], b[l-5], b[l-6])
}

func newConnectedEvent(b []byte) Event {
	eb, l := newEventBase(b)
	return &ConnectedEvent{
		eb,
		macString(b, l),
	}
}

// DisconnectedEvent structure
type DisconnectedEvent struct {
	*EventBase
	Mac string
}

func newDisconnectedEvent(b []byte) Event {
	eb, l := newEventBase(b)
	return &DisconnectedEvent{
		eb,
		macString(b, l),
	}
}

// VibrationEvent structure
type VibrationEvent struct {
	*EventBase
}

func newVibrationEvent(b []byte) Event {
	eb, _ := newEventBase(b)
	return &VibrationEvent{
		eb,
	}
}
