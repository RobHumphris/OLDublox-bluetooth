package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	"github.com/RobHumphris/rf-gateway/global"
	u "github.com/RobHumphris/ublox-bluetooth"
	retry "github.com/avast/retry-go"

	"github.com/pkg/errors"
)

func TestExtendedDataMode(t *testing.T) {
	ub, err := u.NewUbloxBluetooth("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("NewUbloxBluetooth error %v\n", err)
	}

	err = ub.EnterExtendedDataMode()
	if err != nil {
		t.Fatalf("EnterDataMode error %v\n", err)
	}

	err = ub.ATCommand()
	if err != nil {
		t.Fatalf("AT Command 1 error %v\n", err)
	}
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 1000; i++ {
		e := retry.Do(func() error {
			return workflowTest("D5926479C652r", i, ub)
		},
			retry.Attempts(global.RetryCount),
			retry.Delay(global.RetryWait))
		if e != nil {
			err = errors.Wrapf(err, "workflowTest error %v", e)
			t.Fatalf("\nWorkflow test error %v\n", err)
			break
		}
	}

	if err != nil {
		t.Fatalf("Workflow test errors %v\n", err)
	}
}

func workflowTest(mac string, itteration int, ub *u.UbloxBluetooth) error {
	err := ub.ConnectToDevice(mac,
		func() error {
			fmt.Printf("Workflow Test run: %d\n", itteration)
			err := ub.EnableNotifications()
			if err != nil {
				return errors.Wrap(err, "EnableNotifications error")
			}

			err = ub.EnableIndications()
			if err != nil {
				return errors.Wrap(err, "EnableIndications error %v\n")
			}

			unlocked, err := ub.UnlockDevice(password)
			if err != nil {
				return errors.Wrap(err, "UnlockDevice error")
			}
			if !unlocked {
				return fmt.Errorf("UnlockDevice error - failed to unlock")
			}

			return ub.DisconnectFromDevice()
		},
		func() error {
			fmt.Printf("Unexpected disconnect")
			return fmt.Errorf("Unexpected disconnect")
		})

	return err
}
