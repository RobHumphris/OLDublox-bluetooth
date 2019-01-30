package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"
)

var timeout = 6 * time.Second

//sudo interceptty /dev/ttyUSB0
func TestFunctionality(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", 1*time.Second)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = testATCommand(ub, t)
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}
	time.Sleep(1 * time.Second)

	//ub.Write("ATO2")
	//err = ub.WaitForResponse(nil, (6 * time.Second))
	//if err != nil {
	//	t.Errorf("ATO2 error %v\n", err)
	//}

	err = testDiscovery(ub, t)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}

	time.Sleep(5 * time.Second)
	/*err = testDiscovery(ub, t)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}*/

	time.Sleep(1 * time.Second)
	/*err = testConnect(ub, t)
	if err != nil {
		t.Errorf("TestConnect error %v\n", err)
	}
	time.Sleep(1 * time.Second)*/
	//ub.Close()
}

func testATCommand(ub *UbloxBluetooth, t *testing.T) error {
	atFn := func(d DataResponse) error {
		fmt.Printf("--> %v\n", d)
		return nil
	}

	ub.Write("AT")
	return ub.WaitForResponse(atFn, timeout)
}

func testDiscovery(ub *UbloxBluetooth, t *testing.T) error {
	err := ub.Write(DiscoveryCommand())
	if err != nil {
		t.Fatalf("Write error %v\n", err)
	}

	discFn := func(d DataResponse) error {
		if d.token == discovery {
			discovered, err := NewDiscoveryReply(string(d.data[:]))
			fmt.Printf("Descovered device: %v\n", discovered)
			return err
		}
		return fmt.Errorf("Incorrect token %s for DiscoveryReply", d.token)
	}

	err = ub.WaitForResponse(discFn, timeout)
	return err
}

func testConnect(ub *UbloxBluetooth, t *testing.T) error {
	err := ub.Write(ConnectCommand("C1851F6083F8r"))
	if err != nil {
		t.Fatalf("Write error %v\n", err)
	}

	dataFn := func(d DataResponse) error {
		fmt.Printf("Connect response token: %s data: %s\n", d.token, string(d.data[:]))
		return nil
	}

	err = ub.WaitForResponse(dataFn, timeout)
	return err
}
