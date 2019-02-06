package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

const serviceUUIDLength = 42
const serviceUUIDHeaderLength = 10
const serviceUUID = "23E1B7EA5F782315A7BEADDE10138888"

// TestUbloxBluetoothCommands treads through the list of implemented commands
func TestUbloxBluetoothCommands(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
	time.Sleep(1 * time.Second)

	discovered, err := ub.DiscoveryCommand()
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}
	if len(discovered) < 1 {
		t.Errorf("No discovered devices found\n")
	}

	for _, device := range discovered {
		if len(device.Data) != serviceUUIDLength || device.Data[serviceUUIDHeaderLength:] != serviceUUID {
			fmt.Printf("Not Testing %s\n", device.BluetoothAddress)
		} else {
			fmt.Printf("Testing device %s\n", device.BluetoothAddress)
			exerciseTheDevice(device.BluetoothAddress, ub, t)
		}
	}
}

func exerciseTheDevice(deviceAddr string, ub *UbloxBluetooth, t *testing.T) {
	cr, err := ub.ConnectToDevice(deviceAddr)

	fmt.Printf("[ConnectToDevice] replied with: %v\n", cr)
	if err != nil {
		t.Errorf("TestConnect error %v\n", err)
	}
	if cr.BluetoothAddress != deviceAddr {
		t.Errorf("ConnectToDevice - addresses do not match")
	}
	if cr.Type != 0 {
		t.Errorf("ConnectToDevice - type is unknown should be zero")
	}
	time.Sleep(1 * time.Second)

	err = ub.EnableIndications(cr)
	if err != nil {
		t.Errorf("EnableIndications error %v\n", err)
	}

	err = ub.EnableNotifications(cr)
	if err != nil {
		t.Errorf("EnableNotifications error %v\n", err)
	}

	unlocked, err := ub.UnlockDevice(cr, password)
	if err != nil {
		t.Errorf("UnlockDevice error %v\n", err)
	}
	if !unlocked {
		t.Errorf("UnlockDevice error - failed to unlock")
	}
	fmt.Printf("[UnlockDevice] replied with: %v\n", unlocked)

	version, err := ub.GetVersion(cr)
	if err != nil {
		t.Errorf("GetVersion error %v\n", err)
	}
	fmt.Printf("[GetVersion] replied with: %v\n", version)

	info, err := ub.GetInfo(cr)
	if err != nil {
		t.Errorf("GetInfo error %v\n", err)
	}
	fmt.Printf("[GetInfo] replied with: %v\n", info)

	config, err := ub.ReadConfig(cr)
	if err != nil {
		t.Errorf("ReadConfig error %v\n", err)
	}
	fmt.Printf("[ReadConfig] replied with: %v\n", config)

	startingIndex := info.CurrentSequenceNumber - info.RecordsCount
	fmt.Printf("[DownloadLogFile] starting")
	_, err = ub.DownloadLogFile(cr, startingIndex)
	if err != nil {
		t.Errorf("DownloadLogFile error %v\n", err)
	}
	fmt.Printf("[DownloadLogFile] complete")

	/*
		slotCount, err := ub.ReadSlotCount(cr)
		if err != nil {
			t.Errorf("ReadSlotCount error %v\n", err)
		}
		fmt.Printf("[ReadSlotCount] replied with: %v\n", slotCount)

		slotInfo, err := ub.ReadSlotInfo(cr, 1)
		if err != nil {
			t.Errorf("ReadSlotInfo error %v\n", err)
		}
		fmt.Printf("[ReadSlotInfo] replied with: %v\n", slotInfo)

		slotData, err := ub.ReadSlotData(cr, 1, 0)
		if err != nil {
			t.Errorf("ReadSlotData error %v\n", err)
		}
		fmt.Printf("[ReadSlotData] replied with: %v\n", slotData)
	*/

	err = ub.DisconnectFromDevice(cr)
	if err != nil {
		t.Errorf("DisconnectFromDevice error %v\n", err)
	}
}
