package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

func newBE(i uint32) *VehEvent {
	return &VehEvent{
		DataFlag:  true,
		Sequence:  i,
		Timestamp: uint32(time.Now().Unix()),
		EventType: VehEventBoot,
		BootEvent: &VehBootEvent{
			Reason:          i,
			SoftwareVersion: fmt.Sprintf("S/W %d", i),
			HardwareVersion: i,
			BuildNumber:     i,
		},
	}
}

func mapEvent(e *VehEvent) {
	switch e.EventType {
	case VehEventBoot:
		fmt.Printf("BootEvent %d\n", e.Sequence)
	case VehEventSensor:
		fmt.Printf("SensorEvent %d\n", e.Sequence)
	case VehEventConnected:
		fmt.Printf("ConnectedEvent %d\n", e.Sequence)
	case VehEventDisconnected:
		fmt.Printf("DisconnectedEvent %d\n", e.Sequence)
	case VehEventVibration:
		fmt.Printf("VibrationEvent %d\n", e.Sequence)
	case VehEventError:
		fmt.Printf("ErrorEvent %d\n", e.Sequence)
	default:
		fmt.Println("NON!")
	}
}

func TestPackingAndCasting(t *testing.T) {
	events := []*VehEvent{}

	for i := 0; i < 10; i++ {
		events = append(events, newBE(uint32(i)))
	}

	for _, e := range events {
		mapEvent(e)
	}
}
