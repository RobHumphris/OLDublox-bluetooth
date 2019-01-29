package ubloxbluetooth

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/RobHumphris/ublox-bluetooth/serial"
)

// DataResponse holds the Token at the start of the reply, and the subsequent data bytes
type DataResponse struct {
	token string
	data  []byte
}

// UbloxBluetooth holds the serial port, and the communication channels.
type UbloxBluetooth struct {
	lastCommand string
	serialPort  *serial.SerialPort
	readChannel chan []byte
	//DiscoveryChannel chan *DiscoveryReply
	DataChannel      chan DataResponse
	ErrorChannel     chan error
	CompletedChannel chan bool
}

// DiscoveryFunc is called when the UbloxBluetooth DiscoveryChannel receives a message
//type DiscoveryFunc func(*DiscoveryReply) error

// DataFunc is called when the UbloxBluetooth DataChannel recieves a message
type DataFunc func(DataResponse) error

// NewUbloxBluetooth creates a new UbloxBluetooth instance
func NewUbloxBluetooth(device string, timeout time.Duration) (*UbloxBluetooth, error) {
	sp, err := serial.OpenSerialPort(device, timeout)
	if err != nil {
		return nil, err
	}

	ub := &UbloxBluetooth{
		lastCommand: "",
		serialPort:  sp,
		readChannel: make(chan []byte),
		//DiscoveryChannel: make(chan *DiscoveryReply),
		DataChannel:      make(chan DataResponse),
		ErrorChannel:     make(chan error),
		CompletedChannel: make(chan bool),
	}
	sp.Flush()

	go ub.serialportReader()

	return ub, nil
}

var tail = []byte("\r\n")

// Write writes the data string to Ublox via the SerialPort
func (ub *UbloxBluetooth) Write(data string) error {
	fmt.Printf("Writing %s to u-blox\n", data)
	ub.lastCommand = data
	return ub.serialPort.Write([]byte(append([]byte(data), tail...)))
}

// WaitForResponse waits until timeout for a response from
//func (ub *UbloxBluetooth) WaitForResponse(disFn DiscoveryFunc, datFn DataFunc, timeout time.Duration) error {
func (ub *UbloxBluetooth) WaitForResponse(datFn DataFunc, timeout time.Duration) error {
	var err error
	loop := true
	for loop {
		select {
		/*case disCh := <-ub.DiscoveryChannel:
		if disFn != nil {
			err = disFn(disCh)
		} else {
			err = fmt.Errorf("No DiscoveryFunc defined")
		}*/
		case datCh := <-ub.DataChannel:
			if datFn != nil {
				err = datFn(datCh)
			} else {
				err = fmt.Errorf("No DataFunc defined")
			}
		case _ = <-ub.CompletedChannel:
			err = nil
			loop = false
		case e := <-ub.ErrorChannel:
			err = e
			loop = false
		case <-time.After(timeout):
			err = fmt.Errorf("Timeout")
			loop = false
		}
	}
	return err
}

// Close shuts down the serial port, can closes communication channels.
func (ub *UbloxBluetooth) Close() {
	ub.Close()
	close(ub.readChannel)
	//close(ub.DiscoveryChannel)
	close(ub.ErrorChannel)
}

func (ub *UbloxBluetooth) serialportReader() {
	fmt.Println("[serialportReader] Start")
	defer fmt.Println("[serialportReader] End")
	go ub.serialPort.ScanLines(ub.readChannel)
	for {
		b := <-ub.readChannel
		if len(b) != 0 {
			str := string(b[:])
			switch b[0] {
			case 'A':
				ub.processATCommands(str)
			case '+':
				ub.processCommandResponse(b)
			default:
				ub.handleMessage(str)
			}
		}
	}
}

var separator = []byte(":")
var empty = ""

func (ub *UbloxBluetooth) processATCommands(str string) {
	fmt.Println("[processATCommands] Read:", str)
	if strings.HasPrefix(str, "AT+") {
		if ub.lastCommand != empty {
			if ub.lastCommand == str {
				fmt.Printf("Command %s echoed\n", str)
				ub.lastCommand = empty
				return
			}
		}
		ub.ErrorChannel <- fmt.Errorf("unexpected reply %s", str)
	}
}

func (ub *UbloxBluetooth) processCommandResponse(b []byte) {
	//fmt.Println("[processCommandResponse] Read:", string(b[:]))
	d := bytes.Split(b, separator)
	resp := DataResponse{
		token: string(d[0][:]),
		data:  d[1],
	}
	ub.DataChannel <- resp
}

func (ub *UbloxBluetooth) handleMessage(p string) {
	switch p {
	case okMessage:
		ub.CompletedChannel <- true
	case errorMessage:
		ub.ErrorChannel <- fmt.Errorf(p)
	default:
		ub.ErrorChannel <- fmt.Errorf("Cannot handle message %s", p)
	}
}

func (ub *UbloxBluetooth) handleUnknownPayload(t string, p string) {
	ub.ErrorChannel <- fmt.Errorf("Unknown token %s payload %s", t, p)
}
