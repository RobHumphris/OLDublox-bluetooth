package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	ub "github.com/RobHumphris/ublox-bluetooth"
	serial "github.com/RobHumphris/ublox-bluetooth/serial"
)

func setupForSerialTests(t *testing.T, echoOff bool) (*ub.UbloxBluetooth, error) {
	ub, err := ub.NewUbloxBluetooth(timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}

	if echoOff {
		err = ub.EchoOff()
		if err != nil {
			t.Errorf("EchoOff error %v\n", err)
		}
	}

	return ub, err
}

func TestReset(t *testing.T) {
	ub, err := setupForSerialTests(t, false)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}

	err = ub.ResetSerial()
	if err != nil {
		t.Fatalf("ResetSerial error %v\n", err)
	}

	time.Sleep(5 * time.Second)

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}

	ub.Close()
}

func TestDataMode(t *testing.T) {
	ub, err := setupForSerialTests(t, false)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.EnterDataMode()
	if err != nil {
		t.Fatalf("EnterDataMode error %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)

	err = ub.EnterCommandMode()
	if err != nil {
		t.Fatalf("EnterCommandMode error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command error %v\n", err)
	}
}

func TestSerialPortService(t *testing.T) {
	loopCount := 0
	ub, err := setupForSerialTests(t, true)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	for {
		fmt.Printf("%s Loop count %d: ", time.Now().String(), loopCount)
		loopCount++

		fmt.Print("A")
		serial.SetVerbose(true)
		h, err := ub.ConnectDeviceSPS("D4CA6EBE5AC8p")
		if err != nil {
			t.Fatalf("EnableSerialPort error %v\n", err)
		}

		time.Sleep(50 * time.Millisecond)
		fmt.Print("B")
		err = ub.EnterDataMode()
		if err != nil {
			t.Errorf("EnterDataMode error %v\n", err)
		}

		time.Sleep(50 * time.Millisecond)

		fmt.Print("C")
		/*for i := 0; i < 1000; i++ {
			err = ub.WriteSPS([]byte("onetwothreefourfivesix"))
			if err != nil {
				t.Errorf("WriteSPS error %v\n", err)
			}
			time.Sleep(50 * time.Millisecond)
		}*/

		fmt.Print("D")
		err = ub.DisconnectDeviceSPS(h)
		if err != nil {
			t.Errorf("DisconnectDeviceSPS error %v\n", err)
		}

		fmt.Print("E")
		err = ub.ATCommand()
		if err != nil {
			t.Fatalf("AT Command error %v\n", err)
		}
		time.Sleep(100 * time.Millisecond)
		fmt.Println()
	}
}
