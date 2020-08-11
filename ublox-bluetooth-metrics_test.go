package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
)

func TestRSSICommand(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	ub.serialPort.SetVerbose(true)
	doRSSITest(os.Getenv("DEVICE_MAC"), ub, t)

}

func doRSSITest(address string, ub *UbloxBluetooth, t *testing.T) {
	err := connectToDevice(address, func(t *testing.T) error {
		rssi, er := ub.GetRSSI()
		if er != nil {
			return er
		}

		fmt.Printf("RSSI Channel: %d dbm: %d\n", rssi.Channel, rssi.Dbm)
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("Connect to device error %v\n", err)
	}
}
