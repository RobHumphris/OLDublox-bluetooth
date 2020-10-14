package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/fortytw2/leaktest"
	"github.com/pkg/errors"
)

func TestMultipleConnects(t *testing.T) {
	defer leaktest.Check(t)()
	btd, err := InitUbloxBluetooth(timeout, testErrorHandler)
	if err != nil {
		t.Fatalf("InitUbloxBluetooth error %v", err)
	}

	ub, err := btd.GetDevice(0)

	//defer ub.Close()
	ub.serialPort.SetVerbose(true)

	err = ub.ConfigureUblox(timeout)
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

	fmt.Printf("Starting connect test ")
	for i := 0; i < 100; i++ {
		err = retry.Do(func() error {
			fmt.Printf("%03d ", i)
			e := doConnect(ub, os.Getenv("DEVICE_MAC"), i)
			if e != nil {
				fmt.Printf("doConnect error %v", err)
			}
			return e
		},
			retry.Attempts(3),
			retry.Delay(500*time.Millisecond))
	}
	ub.Close()
}

func doConnect(ub *UbloxBluetooth, mac string, count int) error {
	st := time.Now().UnixNano()
	err := ub.ConnectToDevice(mac, func(ub *UbloxBluetooth) error {
		defer ub.DisconnectFromDevice()
		tt := time.Now().UnixNano() - st
		fmt.Printf("Connection delay: %dns\n", tt)

		time.Sleep(20 * time.Millisecond)

		err := ub.EnableNotifications()
		if err != nil {
			return errors.Wrapf(err, "EnableNotifications mac: %s count: %d\n", mac, count)
		}

		err = ub.EnableIndications()
		if err != nil {
			return errors.Wrapf(err, "EnableIndications mac: %s count: %d\n", mac, count)
		}

		_, err = ub.UnlockDevice(password)
		if err != nil {
			return errors.Wrapf(err, "UnlockDevice mac: %s count: %d\n", mac, count)
		}
		return nil
	}, func(ub *UbloxBluetooth) error {
		return fmt.Errorf("Disconnected")
	})
	if err != nil {
		return errors.Wrapf(err, "TestConnect mac: %s count: %d\n", mac, count)
	}
	return err
}
