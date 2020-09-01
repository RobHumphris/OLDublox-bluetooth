package ubloxbluetooth

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestTimeFunctions(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice(os.Getenv("DEVICE_MAC"), func(t *testing.T) error {
		currentTime := int32(time.Now().Unix())
		deviceTime, err := ub.GetTime()
		if err != nil {
			return errors.Wrap(err, "GetTime error")
		}

		fmt.Printf("Current time: %d, device time: %d\n", currentTime, deviceTime)

		timeAdjust, err := ub.SetTime(currentTime)
		if err != nil {
			return errors.Wrap(err, "SetTime error")
		}

		fmt.Printf("TimeAdjustReply CurrentTime: %d UpdatedTime: %d\n", timeAdjust.CurrentTime, timeAdjust.UpdatedTime)
		return nil
	}, ub, t)

	if err != nil {
		t.Errorf("Connect to device error %v\n", err)
	}
}
