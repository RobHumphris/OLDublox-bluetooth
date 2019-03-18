package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	"github.com/RobHumphris/rf-gateway/global"
	u "github.com/RobHumphris/ublox-bluetooth"
	serial "github.com/RobHumphris/ublox-bluetooth/serial"
	retry "github.com/avast/retry-go"
	"github.com/pkg/errors"
)

func TestMultipleConnects(t *testing.T) {
	serial.SetVerbose(true)
	ub, err := u.NewUbloxBluetooth(timeout)
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
			//e := doConnect(ub, "C1851F6083F8r", i)
			e := doConnect(ub, "CE1A0B7E9D79r", i)
			if e != nil {
				fmt.Printf("doConnect error %v", err)
			}
			return e
		},
			retry.Attempts(global.RetryCount),
			retry.Delay(global.RetryWait))

		//doConnect(ub, "CE1A0B7E9D79r", t)
		//doConnect(ub, "D8CFDFA118ECr", t)
	}
}

func doConnect(ub *u.UbloxBluetooth, mac string, count int) error {
	fmt.Print("C")
	err := ub.ConnectToDevice(mac, func() error {
		defer ub.DisconnectFromDevice()

		time.Sleep(global.BluetoothPostConnectDelay)

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
