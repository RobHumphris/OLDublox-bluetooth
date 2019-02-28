package ubloxbluetooth

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestExtendedDataMode(t *testing.T) {
	ub, err := NewUbloxBluetooth("/dev/ttyUSB0", timeout)
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
		e := workflowTest("D5926479C652r", i, ub)
		if e != nil {
			err = errors.Wrapf(err, "workflowTest error %v", e)
			fmt.Printf("\nWorkflow test error %v\n", e)
			break
		}
		//exerciseTheDevice("CE1A0B7E9D79r", ub, t, i, true, false)
		//exerciseTheDevice("D5926479C652r", ub, t, i, false, false)
		//time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("Workflow test errors %v\n", err)
	}
}

func workflowTest(mac string, itteration int, ub *UbloxBluetooth) error {
	err := ub.ConnectToDevice(mac,
		func(cr *ConnectionReply) error {
			fmt.Printf("Workflow Test run: %d\n", itteration)
			err := ub.EnableNotifications(cr)
			if err != nil {
				return errors.Wrap(err, "EnableNotifications error")
			}

			err = ub.EnableIndications(cr)
			if err != nil {
				return errors.Wrap(err, "EnableIndications error %v\n")
			}

			unlocked, err := ub.UnlockDevice(cr, password)
			if err != nil {
				return errors.Wrap(err, "UnlockDevice error")
			}
			if !unlocked {
				return fmt.Errorf("UnlockDevice error - failed to unlock")
			}
			/*
				info, err := ub.GetInfo(cr)
				if err != nil {
					return errors.Wrap(err, "GetInfo error %v\n")
				}
				fmt.Printf("[GetInfo] replied with: %v\n", info)

				version, err := ub.GetVersion(cr)
				if err != nil {
					return errors.Wrap(err, "GetVersion error")
				}
				fmt.Printf("[GetVersion] replied with: %v\n", version)

				config, err := ub.ReadConfig(cr)
				if err != nil {
					return errors.Wrap(err, "ReadConfig error")
				}
				fmt.Printf("[ReadConfig] replied with: %v\n", config)

				startingIndex := info.CurrentSequenceNumber - info.RecordsCount
				fmt.Printf("[DownloadLogFile] starting run: %d\n", itteration)
				err = ub.DownloadLogFile(cr, startingIndex, func(b []byte) error {
					//fmt.Print(".")
					return nil
				})
				if err != nil {
					return errors.Wrap(err, "DownloadLogFile error")
				}
				fmt.Printf("[DownloadLogFile] complete\n")
			*/
			return ub.DisconnectFromDevice(cr)
		},
		func(cr *ConnectionReply) error {
			fmt.Printf("Unexpected disconnect")
			return fmt.Errorf("Unexpected disconnect")
		})

	return err
}
