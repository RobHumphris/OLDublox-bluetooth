package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestWriteMessage(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		ub.serialPort.SetVerbose(true)
		e := ub.WriteMessage(fmt.Sprintf("Message at %d", time.Now().Unix()))
		if e != nil {
			return errors.Wrap(e, "WriteMessage error")
		}

		e = ub.DisconnectFromDevice()
		if e != nil {
			return errors.Wrap(e, "DisconnectFromDevice error")
		}
		return nil
	}, ub, t)

	if err != nil {
		t.Errorf("TestWriteMessage error %v\n", err)
	}
}
