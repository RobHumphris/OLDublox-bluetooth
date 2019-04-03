package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	ub "github.com/RobHumphris/ublox-bluetooth"
	"github.com/RobHumphris/ublox-bluetooth/serial"
)

func main() {
	reset := flag.Bool("reset", false, "Reset the device")
	flag.Parse()

	bt, err := ub.NewUbloxBluetooth(6 * time.Second)
	if err != nil {
		log.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer bt.Close()

	if *reset {
		fmt.Println("Resetting device")
		factoryReset(bt)
	} else {
		fmt.Println("Setting defaults")
		defaultSettings(bt)
	}
}

func factoryReset(bt *ub.UbloxBluetooth) {
	serial.SetVerbose(true)
	err := bt.MultipleATCommands()
	if err != nil {
		log.Fatalf("NewUbloxBluetooth AT - 0 error %v\n", err)
		return
	}
	fmt.Println("FactoryReset AT response OK")

	err = bt.FactoryReset()
	if err != nil {
		fmt.Printf("FactoryReset error %v\n", err)
		return
	}
	fmt.Println("FactoryReset OK")

	err = bt.RebootUblox()
	if err != nil {
		fmt.Printf("FactoryReset reboot error %v\n", err)
		return
	}
	fmt.Println("FactoryReset reboot OK")

	err = bt.SetCommsRate(serial.Default)
	if err != nil {
		fmt.Printf("Baud Rate reset error %v\n", err)
		return
	}

	err = bt.MultipleATCommands()
	if err != nil {
		log.Fatalf("NewUbloxBluetooth AT error %v\n", err)
		return
	}
	fmt.Println("FactoryReset AT - 1 response OK")
}

func defaultSettings(bt *ub.UbloxBluetooth) {
	/*err := bt.SetCommsRate(serial.Default)
	if err != nil {
		fmt.Printf("Baud Rate reset error %v\n", err)
		return
	}*/

	serial.SetVerbose(true)

	err := bt.MultipleATCommands()
	if err != nil {
		log.Fatalf("NewUbloxBluetooth AT error %v\n", err)
		return
	}
	fmt.Println("FactoryReset AT - 0 response OK")

	err = bt.ConfigureUblox()
	if err != nil {
		fmt.Printf("ConfigureUblox error %v\n", err)
		return
	}
	fmt.Println("ConfigureUblox OK")

	err = bt.SetRS232BaudRate(1000000)
	if err != nil {
		fmt.Printf("SetRS232BaudRate error %v\n", err)
		return
	} else {
		fmt.Println("SetRS232BaudRate OK")
	}

	err = bt.RebootUblox()
	if err != nil {
		fmt.Printf("ConfigureUblox reboot error %v\n", err)
		return
	}
	fmt.Println("ConfigureUblox reboot OK")

	err = bt.SetModuleStartMode(ub.ExtendedDataMode)
	if err != nil {
		fmt.Printf("SetModuleStartMode error %v\n", err)
	}
	fmt.Println("SetModuleStartMode OK")

	err = bt.SetWatchdogConfiguration()
	if err != nil {
		fmt.Printf("SetWatchdogConfiguration error %v\n", err)
	}
	fmt.Println("SetWatchdogConfiguration OK")

	err = bt.SetCommsRate(serial.HighSpeed)
	if err != nil {
		fmt.Printf("Baud Rate reset error %v\n", err)
		return
	}

	err = bt.RebootUblox()
	if err != nil {
		fmt.Printf("ConfigureUblox reboot error %v\n", err)
		return
	}
	fmt.Println("ConfigureUblox reboot OK")

	err = bt.MultipleATCommands()
	if err != nil {
		log.Fatalf("NewUbloxBluetooth AT error %v\n", err)
		return
	}
	fmt.Println("FactoryReset AT - 2 response OK")
}
