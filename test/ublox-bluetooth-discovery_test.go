package ubloxbluetooth

import (
	"fmt"
	"testing"

	u "github.com/RobHumphris/ublox-bluetooth"
)

// TestDiscovery
func TestDiscovery(t *testing.T) {
	ub, err := u.NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = ub.EnterExtendedDataMode()
	if err != nil {
		t.Fatalf("EnterDataMode error %v\n", err)
	}

	/*err = ub.ConfigureUblox()
	if err != nil {
		t.Fatalf("ConfigureUblox error %v\n", err)
	}

	err = ub.RebootUblox()
	if err != nil {
		t.Fatalf("RebootUblox error %v\n", err)
	}*/

	err = ub.ATCommand()
	if err != nil {
		t.Errorf("AT error %v\n", err)
	}

	alpha := func(dr *u.DiscoveryReply) error {
		fmt.Printf("Discovery: %v\n", dr)
		return nil
	}

	err = ub.DiscoveryCommand(alpha)
	if err != nil {
		t.Errorf("TestDiscovery error %v\n", err)
	}

	/*err = connectToDevice("D5926479C652r", func(cr *ConnectionReply, t *testing.T) error {
		ub.DisconnectFromDevice(cr)

		err = ub.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}

		err = ub.DiscoveryCommand(alpha)
		if err != nil {
			t.Errorf("TestDiscovery error %v\n", err)
		}

		err = ub.ATCommand()
		if err != nil {
			t.Errorf("AT error %v\n", err)
		}
		return err
	}, ub, t)*/

}
