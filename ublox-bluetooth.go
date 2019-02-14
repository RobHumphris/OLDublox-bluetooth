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

	err = sp.Flush()
	if err != nil {
		sp.Close()
		return nil, err
	}

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
	//fmt.Printf("[Write] %s\n", data)
	ub.lastCommand = data
	return ub.serialPort.Write([]byte(append([]byte(data), tail...)))
}

// WaitForResponse waits until timeout for a response from the Ublox device
func (ub *UbloxBluetooth) WaitForResponse(expectedResponse string, waitForData bool) ([]byte, error) {
	expected := []byte(expectedResponse)
	d := []byte{}
	complete := false
	dataReceived := false
	for {
		select {
		case data := <-ub.DataChannel:
			if bytes.HasPrefix(data, expected) {
				d = append(d, data...)
				dataReceived = true
				if complete {
					return d, nil
				}
			} else {
				err := handleUnsolicitedMessage(data)
				if err != nil {
					return d, err
				}
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

func handleUnsolicitedMessage(data []byte) error {
	if bytes.HasPrefix(data, ubloxBTReponseHeader) {
		// Todo - handle the likes of +UUBTLEPHYU:0,0,2,2
	} else {
		if bytes.HasPrefix(data, rebootResponse) {
			return fmt.Errorf("Error device has rebooted")
		}
		fmt.Printf("**** [handleUnexpectedMessage] %s ****\n", data)
	}
	return nil
}

type downloadhandler func([]byte) (bool, error)

// HandleDataDownload handles data notifications
func (ub *UbloxBluetooth) HandleDataDownload(expected int, fn downloadhandler) (int, error) {
	var err error
	received := 0
	loop := true
	indicationRcvd := false
	for {
		select {
		case data := <-ub.DataChannel:
			if received < expected {
				received++
				loop, err = fn(data)
			} else {
				// now waiting for +UUBTGI:0,13,0700
				if bytes.HasPrefix(data, gattIndicationResponse) {
					_, e := splitOutResponse(data, readEventLogReply)
					if e != nil {
						return received, err
					}
					indicationRcvd = true
				}
			}
			if (received == expected && indicationRcvd) || !loop {
				return received, err
			}
		case <-time.After(ub.timeout):
			return received, fmt.Errorf("Timeout")
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
	str := string(b[:])
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
	str := string(b[:])
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
