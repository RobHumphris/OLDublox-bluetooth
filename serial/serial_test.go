package serial

import (
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSerial(t *testing.T) {
	device := "/dev/ttyUSB0"
	timeout := 5 * time.Second
	readChannel := make(chan []byte)
	sp, err := OpenSerialPort(device, timeout)
	if err != nil {
		t.Fatalf("Open Port Error %v\n", err)
	}
	sp.Flush()

	go sp.ScanLines(readChannel)
	go func() {
		for {
			s := <-readChannel
			fmt.Println(s)
		}
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, _ := reader.ReadString('\n')
			sp.Write([]byte(line))
		}
	}()

	select {}
}
