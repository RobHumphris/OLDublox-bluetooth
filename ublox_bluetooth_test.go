package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", 6*time.Second)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.Write(Discovery)
	if err != nil {
		t.Fatalf("Write error %v\n", err)
	}

	loop := true
	for loop {
		select {
		case discovered := <-ub.discoveryChannel:
			fmt.Printf("Descovered device: %v\n", discovered)
		case err := <-ub.errorChannel:
			fmt.Printf("Error recieved: %v\n", err)
		case <-time.After(6 * time.Second):
			fmt.Println("timeout")
			loop = false
		}
	}

}
