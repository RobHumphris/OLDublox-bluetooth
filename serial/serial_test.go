package serial

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSerial(t *testing.T) {
	timeout := 10 * time.Second
	sp, err := OpenSerialPort("/dev/ttyUSB0", timeout)
	if err != nil {
		t.Fatalf("Open Port Error %v\n", err)
	}
	defer func() {
		fmt.Println("Closing serial port")
		err = sp.Close()
		if err != nil {
			t.Fatalf("Close error %v\n", err)
		}
	}()

	sp.Flush()

	err = sp.ResetViaDTR()
	if err != nil {
		t.Fatalf("ResetViaDTR error %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ec := make(chan error, 1)
	go sp.ScanPort(ctx, ec,
		func(b []byte) {
			fmt.Printf("r: %s\n", b)
		}, func(b []byte) {
			fmt.Printf("e: %s\n", b)
		}, func(err error) {
			fmt.Printf("Error: %v\n", err)
		})

	go func() {
		for {
			select {
			case e := <-ec:
				fmt.Printf("Done %v", e)
				return
			}
		}
	}()

	time.Sleep(timeout)
	cancel()
}
