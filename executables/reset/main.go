package main

import (
	"fmt"
	"os"
	"time"

	"github.com/RobHumphris/ublox-bluetooth/serial"
)

func main() {
	sp, err := serial.OpenSerialPort(3 * time.Second)
	if err != nil {
		fmt.Printf("Could not open serial port. Error: %v.\n", err)
		os.Exit(-1)
	}

	err = sp.ResetViaDTR()
	if err != nil {
		fmt.Printf("Reset executed with err: %v.\n", err)
	} else {
		fmt.Println("Reset executed OK.")
	}

	err = sp.Close()
	if err != nil {
		fmt.Printf("Serial Port closed with error: %v. Exiting...\n", err)
	} else {
		fmt.Println("Serial Port closed OK.")
	}

}
