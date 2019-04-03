package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	ub "github.com/RobHumphris/ublox-bluetooth"
	"github.com/RobHumphris/ublox-bluetooth/serial"
)

var password = []byte{'A', 'B', 'C'}

func main() {
	mac := flag.String("mac", "", "MAC of the device")
	flag.Parse()

	fmt.Printf("Opening device MAC %s\n", *mac)

	bt, err := ub.NewUbloxBluetooth(6 * time.Second)
	if err != nil {
		log.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer bt.Close()

	bt.ConnectToDevice(*mac, func() error {
		defer bt.DisconnectFromDevice()

		err := bt.EnableIndications()
		if err != nil {
			return err
		}

		err = bt.EnableNotifications()
		if err != nil {
			return err
		}

		unlocked, err := bt.UnlockDevice(password)
		if err != nil {
			return err
		}
		if !unlocked {
			return fmt.Errorf("Unlock device failed")
		}

		serial.SetVerbose(true)

		info, err := bt.GetInfo()
		if err != nil {
			return err
		}
		fmt.Printf("[GetInfo] replied with: %v\n", info)

		err = bt.ClearEventLog()
		if err != nil {
			return err
		}

		info, err = bt.GetInfo()
		if err != nil {
			return err
		}
		fmt.Printf("[GetInfo] replied with: %v\n", info)

		time.Sleep(5 * time.Second)

		return nil
	}, func() error {
		return fmt.Errorf("unexpected disconnect")
	})

}
