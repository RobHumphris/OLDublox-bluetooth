package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
)

func TestGetVersion(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		v, err := ub.GetVersion()
		if err != nil {
			t.Errorf("GetVersion error %v\n", err)
		}
		fmt.Printf("Software %s Hardware %s Release %s\n", v.SoftwareVersion, v.HardwareVersion, v.ReleaseFlag)
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("TestConfiguration error %v\n", err)
	}
}

func TestConfiguration(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		defer ub.DisconnectFromDevice()

		ub.serialPort.SetVerbose(true)
		cfg, err := ub.ReadConfig()
		if err != nil {
			t.Errorf("ReadConfig error %v\n", err)
		}

		fmt.Printf("Config: %v\n", cfg)
		cfg.SampleTime = cfg.SampleTime + 10

		err = ub.WriteConfig(cfg)
		if err != nil {
			t.Errorf("WriteConfig error %v\n", err)
		}

		ub.serialPort.SetVerbose(false)
		v, err := ub.GetVersion()
		if err != nil {
			t.Errorf("GetVersion error %v\n", err)
		}
		fmt.Printf("Software %s Hardware %s Release %s\n", v.SoftwareVersion, v.HardwareVersion, v.ReleaseFlag)

		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("TestConfiguration error %v\n", err)
	}

}
