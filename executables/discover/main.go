package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	ub "github.com/RobHumphris/ublox-bluetooth"
	"github.com/RobHumphris/ublox-bluetooth/serial"
)

const defaultSerialPort = "/dev/ttyUSB0"

func main() {
	port := flag.String("port", defaultSerialPort, "Ublox serial port")
	flag.Parse()

	fmt.Printf("Using SerialPort: %s\n", *port)

	bt, err := ub.NewUbloxBluetooth(*port, 6*time.Second)
	if err != nil {
		log.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer bt.Close()

	serial.SetVerbose(true)
	err = bt.EnterExtendedDataMode()
	if err != nil {
		log.Fatalf("EnterDataMode error %v\n", err)
	}

	/*err = ub.ConfigureUblox()
	if err != nil {
		t.Fatalf("ConfigureUblox error %v\n", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}*/

	err = bt.ATCommand()
	if err != nil {
		log.Fatalf("AT error %v\n", err)
	}

	alpha := func(dr *ub.DiscoveryReply) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	err = bt.DiscoveryCommand(alpha)
	if err != nil {
		log.Fatalf("TestDiscovery error %v\n", err)
	}

}
