package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	serial "github.com/8power/ublox-bluetooth/serial"
	retry "github.com/avast/retry-go"
	"github.com/pkg/errors"
)

func TestMultipleConnects(t *testing.T) {
	serial.SetVerbose(true)
	ub, err := NewUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.ConfigureUblox()
	if err != nil {
		t.Fatalf("ConfigureUblox error %v\n", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	err = ub.EchoOff()
	if err != nil {
		t.Errorf("EchoOff error %v\n", err)
	}

	settings, err := ub.GetRS232Settings()
	if err != nil {
		t.Fatalf("GetRS232Settings error %v\n", err)
	}
	fmt.Printf("RS232 settings: %v\n", settings)

	fmt.Printf("Starting connect test ")
	for i := 0; i < 10; i++ {
		err = retry.Do(func() error {
			fmt.Printf("%03d ", i)
			e := doConnect(ub, "EE9EF8BA058Br", i)
			if e != nil {
				fmt.Printf("doConnect error %v", err)
			}
			return e
		},
			retry.Attempts(3),
			retry.Delay(500*time.Millisecond))
	}
}

func doConnect(ub *UbloxBluetooth, mac string, count int) error {
	fmt.Print("C")
	err := ub.ConnectToDevice(mac, func() error {
		defer ub.DisconnectFromDevice()

		time.Sleep(200 * time.Millisecond)

		fmt.Print("N")
		err := ub.EnableNotifications()
		if err != nil {
			return errors.Wrapf(err, "EnableNotifications mac: %s count: %d\n", mac, count)
		}

		fmt.Print("I")
		err = ub.EnableIndications()
		if err != nil {
			return errors.Wrapf(err, "EnableIndications mac: %s count: %d\n", mac, count)
		}

		fmt.Print("U")
		_, err = ub.UnlockDevice(password)
		if err != nil {
			return errors.Wrapf(err, "UnlockDevice mac: %s count: %d\n", mac, count)
		}
		fmt.Print("D\n")
		return nil
	}, func() error {
		return fmt.Errorf("Disconnected")
	})
	if err != nil {
		return errors.Wrapf(err, "TestConnect mac: %s count: %d\n", mac, count)
	}
	return err
}
