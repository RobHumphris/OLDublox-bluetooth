package ubloxbluetooth

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/RobHumphris/ublox-bluetooth/serial"
)

var (
	tail      = []byte{'\r', '\n'}
	newline   = "\r\n"
	separator = []byte(":")
	empty     = ""
)

// DataResponse holds the Token at the start of the reply, and the subsequent data bytes
type DataResponse struct {
	token string
	data  []byte
}

// UbloxBluetooth holds the serial port, and the communication channels.
type UbloxBluetooth struct {
	timeout          time.Duration
	lastCommand      string
	serialPort       *serial.SerialPort
	readChannel      chan []byte
	DataChannel      chan []byte
	ErrorChannel     chan error
	CompletedChannel chan bool
}

// DataHandler is called when the UbloxBluetooth DataChannel recieves a message
type DataHandler func(DataResponse) error

// NewUbloxBluetooth creates a new UbloxBluetooth instance
func NewUbloxBluetooth(device string, timeout time.Duration) (*UbloxBluetooth, error) {
	sp, err := serial.OpenSerialPort(device, timeout)
	if err != nil {
		return nil, err
	}
	sp.Flush()

	ub := &UbloxBluetooth{
		timeout:          timeout,
		lastCommand:      "",
		serialPort:       sp,
		readChannel:      make(chan []byte),
		DataChannel:      make(chan []byte), // make(chan DataResponse),
		ErrorChannel:     make(chan error),
		CompletedChannel: make(chan bool),
	}

	go ub.serialportReader()

	return ub, err
}

// Write writes the data string to Ublox via the SerialPort
func (ub *UbloxBluetooth) Write(data string) error {
	fmt.Printf("Writing %q to u-blox\n", data)
	ub.lastCommand = data
	return ub.serialPort.Write([]byte(append([]byte(data), tail...)))
}

// WaitForResponse waits until timeout for a response from
func (ub *UbloxBluetooth) WaitForResponse(waitForData bool) ([]byte, error) {
	d := []byte{}
	complete := false
	dataReceived := false
	for {
		select {
		case data := <-ub.DataChannel:
			d = append(d, data...)
			dataReceived = true
			if complete {
				return d, nil
			}
		case _ = <-ub.CompletedChannel:
			complete = true
			if waitForData {
				if dataReceived {
					return d, nil
				}
			} else {
				return d, nil
			}
		case e := <-ub.ErrorChannel:
			return nil, e
		case <-time.After(ub.timeout):
			return nil, fmt.Errorf("Timeout")
		}
	}
}

// Close shuts down the serial port, can closes communication channels.
func (ub *UbloxBluetooth) Close() {
	ub.serialPort.Close()

	close(ub.readChannel)
	close(ub.DataChannel)
	close(ub.CompletedChannel)
	close(ub.ErrorChannel)
}

func (ub *UbloxBluetooth) serialportReader() {
	fmt.Println("[serialportReader] Start")
	defer fmt.Println("[serialportReader] End")
	go ub.serialPort.ScanLines(ub.readChannel)
	for {
		b := <-ub.readChannel
		b = bytes.Trim(b, newline)
		if len(b) != 0 {
			switch b[0] {
			case 'A':
				ub.processATCommands(b)
			case '+':
				ub.DataChannel <- b
			default:
				ub.handleGeneralMessage(b)
			}
		}
	}
}

func (ub *UbloxBluetooth) processATCommands(b []byte) {
	str := strings.Trim(string(b[:]), "\r\n")
	if strings.HasPrefix(str, at) {
		if ub.lastCommand != empty {
			if strings.HasPrefix(str, ub.lastCommand) {
				ub.lastCommand = empty
				return
			}
		}
		fmt.Printf("unexpected reply %s\n", str)
	}
}

func (ub *UbloxBluetooth) handleGeneralMessage(b []byte) {
	str := strings.Trim(string(b[:]), "\r\n")
	fmt.Printf("[handleGeneralMessage] Processing %s\n", str)
	switch str {
	case okMessage:
		ub.CompletedChannel <- true
	case errorMessage:
		ub.ErrorChannel <- fmt.Errorf(str)
	default:
		ub.ErrorChannel <- fmt.Errorf("Cannot handle message %q", str)
	}
}

func (ub *UbloxBluetooth) handleUnknownPayload(t string, p string) {
	ub.ErrorChannel <- fmt.Errorf("Unknown token %s payload %s", t, p)
}
