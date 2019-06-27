package ubloxbluetooth

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestAbort(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	// put some events into the sensors logs (assuming connection events are being logged)
	for i := 0; i < 20; i++ {
		fmt.Printf("Connect attempt %d\n", i)
		errr := connectToDevice("EE9EF8BA058Br", func(t *testing.T) error {
			return ub.DisconnectFromDevice()
		}, ub, t)
		if errr != nil {
			err = errors.Wrapf(err, "connectToDevice error: %v", errr)
		}
	}

	if err != nil {
		t.Fatalf("TestAbort error %v\n", err)
	}
}
