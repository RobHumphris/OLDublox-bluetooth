package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	"github.com/RobHumphris/rf-gateway/bluetooth"
)

func TestSlotDownload(t *testing.T) {
	mac := "CCBDE3ECB107r" //"CE1A0B7E9D79r" //"D5926479C652r" // "CCBDE3ECB107r"
	for i := 0; i < 20; i++ {
		sc, err := bluetooth.AccessDeviceSlots(mac, 0, 0)
		if err != nil {
			t.Errorf("AccessDeviceSlots error %v", err)
		} else {
			fmt.Printf("Slot count %v\n", sc)
		}
		time.Sleep(5 * time.Second)
	}
}
