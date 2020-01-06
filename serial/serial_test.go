package serial

import (
	"testing"
	"time"
)

func TestSerial(t *testing.T) {
	timeout := 5 * time.Second
	//readChannel := make(chan []byte)
	sp, err := OpenSerialPort(timeout)
	if err != nil {
		t.Fatalf("Open Port Error %v\n", err)
	}
	sp.Flush()

	err = sp.ResetViaDTR()
	if err != nil {
		t.Fatalf("ResetViaDTR error %v\n", err)
	}

	time.Sleep(timeout)

	err = sp.Close()
	if err != nil {
		t.Fatalf("Close error %v\n", err)
	}
	/*go sp.ScanLines(readChannel)
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

	select {}*/
}
