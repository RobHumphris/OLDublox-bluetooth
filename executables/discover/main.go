package main

import (
	"fmt"
	"log"
	"time"

	ub "github.com/RobHumphris/ublox-bluetooth"
	"github.com/RobHumphris/ublox-bluetooth/serial"
)

func main() {
	bt, err := ub.NewUbloxBluetooth(6 * time.Second)
	if err != nil {
		log.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer bt.Close()

	serial.SetVerbose(true)
	err = bt.EnterExtendedDataMode()
	if err != nil {
		log.Fatalf("EnterDataMode error %v\n", err)
	}

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
