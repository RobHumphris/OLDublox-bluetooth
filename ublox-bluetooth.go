package ubloxbluetooth

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/8power/ublox-bluetooth/serial"
)

var ErrRebooted = fmt.Errorf("Error Ublox has rebooted")
var ErrTimeout = fmt.Errorf("Timeout")

// DataResponse holds the Token at the start of the reply, and the subsequent data bytes
type DataResponse struct {
	token string
	data  []byte
}

// Discoveryhandler is called when the UbloxBluetooth DataChannel recieves a message
type Discoveryhandler func([]byte) (bool, error)

// DataMessageHandler functions are invoked when data is recieved.
type DataMessageHandler func([]byte) (bool, error)

// DeviceEvent functions are called and defined in various events e.g. Connect, Disconnect
type DeviceEvent func() error

type ubloxMode int

const commandMode ubloxMode = 0
const dataMode ubloxMode = 1
const extendedDataMode ubloxMode = 2

// UbloxBluetooth holds the serial port, and the communication channels.
type UbloxBluetooth struct {
	timeout            time.Duration
	lastCommand        string
	serialPort         *serial.SerialPort
	currentMode        ubloxMode
	StartEventReceived bool
	readChannel        chan []byte
	DataChannel        chan []byte
	EDMChannel         chan []byte
	ErrorChannel       chan error
	CompletedChannel   chan bool
	stopScanning       chan bool
	connectedDevice    *ConnectionReply
	disconnectHandler  DeviceEvent
	disconnectCount    int
	disconnectExpected bool
}

// NewUbloxBluetooth creates a new UbloxBluetooth instance
func NewUbloxBluetooth(timeout time.Duration) (*UbloxBluetooth, error) {
	sp, err := serial.OpenSerialPort(timeout)
	if err != nil {
		return nil, err
	}

	err = sp.Flush()
	if err != nil {
		sp.Close()
		return nil, err
	}

	ub := &UbloxBluetooth{
		timeout:            timeout,
		lastCommand:        "",
		serialPort:         sp,
		currentMode:        extendedDataMode,
		StartEventReceived: false,
		readChannel:        make(chan []byte),
		DataChannel:        make(chan []byte), // make(chan DataResponse),
		EDMChannel:         make(chan []byte),
		ErrorChannel:       make(chan error),
		CompletedChannel:   make(chan bool),
		stopScanning:       make(chan bool),
		connectedDevice:    nil,
		disconnectCount:    0}

	sp.SetEDMFlag(true)

	go ub.serialportReader()

	return ub, err
}

func (ub *UbloxBluetooth) serialportReader() {
	go ub.serialPort.ScanPort(ub.readChannel, ub.EDMChannel, ub.ErrorChannel)

	for {
		select {
		case b := <-ub.readChannel:
			b = bytes.Trim(b, newline)
			if len(b) != 0 {
				switch b[0] {
				case 'A':
					ub.processATResponse(b)
				case '+':
					ub.DataChannel <- b
				default:
					ub.handleGeneralMessage(b)
				}
			}
		case edmData := <-ub.EDMChannel:
			if len(edmData) > 0 {
				err := ub.ParseEDMMessage(edmData)
				if err != nil {
					ub.ErrorChannel <- err
				}
			}
		case _ = <-ub.stopScanning:
			ub.serialPort.StopScanning()
			return
		}
	}
}

// ResetSerial stops reading threads and
func (ub *UbloxBluetooth) ResetSerial() error {
	ub.stopScanning <- true
	ub.serialPort.Close()

	sp, err := serial.OpenSerialPort(ub.timeout)
	if err != nil {
		return err
	}

	err = sp.Flush()
	if err != nil {
		sp.Close()
		return err
	}

	err = sp.ResetViaDTR()
	if err != nil {
		sp.Close()
		return err
	}

	ub.serialPort = sp
	go ub.serialportReader()

	return nil
}

// Close shuts down the serial port, can closes communication channels.
func (ub *UbloxBluetooth) Close() {
	err := ub.serialPort.Close()
	if err != nil {
		fmt.Printf("[Close] error %v\n", err)
	}

	close(ub.readChannel)
	close(ub.DataChannel)
	close(ub.EDMChannel)
	close(ub.CompletedChannel)
	close(ub.ErrorChannel)
}

// SetCommsRate sets the rate to either: Default BaudRate, or HighSpeed
func (ub *UbloxBluetooth) SetCommsRate(rate serial.BaudRate) error {
	return ub.serialPort.SetBaudRate(rate, ub.timeout)
}

// SetSerialVerbose sets the debug flag
func (ub *UbloxBluetooth) SetSerialVerbose(f bool) {
	serial.SetVerbose(f)
}

// Write writes the data string to Ublox via the SerialPort
func (ub *UbloxBluetooth) Write(data string) error {
	var b []byte
	ub.lastCommand = data

	if ub.currentMode == extendedDataMode {
		b = NewEDMATCommand(data)
	} else {
		b = []byte(append([]byte(data), tail...))
	}
	return ub.WriteBytes(b)
}

// WriteBytes writes the passed bytes
func (ub *UbloxBluetooth) WriteBytes(b []byte) error {
	return ub.serialPort.Write(b)
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
			return nil, ErrTimeout
		}
	}
}

func handleUnsolicitedMessage(data []byte) error {
	if bytes.HasPrefix(data, ubloxBTReponseHeader) {
		// Todo - handle the likes of +UUBTLEPHYU:0,0,2,2
	} else {
		if bytes.HasPrefix(data, rebootResponse) {
			return ErrRebooted
		}
		fmt.Printf("**** [handleUnexpectedMessage] %s ****\n", data)
	}
	return nil
}

func getPayload(d []byte, sep []byte) []byte {
	s := bytes.Split(d, sep)
	return s[1]
}

var notificationSeperator = []byte("16,")
var indicationSeperator = []byte("13,")

// HandleDataDownload enables data download (Events and Slots). Passed variables are:
// `expected` number of notifications. This handles the credit based flow mechanism and does
// not return until the expected number of notifications and terminating indication are received.
//
// `commandReply` the Veh command (0x07 or 0x10)
//
// `dnh` Notification handler function which is invoked each time a notification is received.
//
// `dih` Indication handler function, which is invoked each time an indication is received.
func (ub *UbloxBluetooth) HandleDataDownload(expected int, commandReply string, dnh func([]byte) error, dih func([]byte) error) error {
	var err error
	received := 0
	dataComplete := false
	indicationRecieved := false

	for {
		select {
		case data := <-ub.DataChannel:
			if bytes.HasPrefix(data, gattNotificationResponse) {
				err = dnh(getPayload(data, notificationSeperator))
				if err != nil {
					return err
				}
				received++
				if received%halfwayPoint == 0 {
					err = ub.SendCredits(halfwayPoint)
					if err != nil {
						return err
					}
				}
				dataComplete = (received == expected)
				if dataComplete && indicationRecieved {
					return nil
				}
			} else if bytes.HasPrefix(data, gattIndicationResponse) {
				err = dih(getPayload(data, indicationSeperator))
				if err != nil {
					return err
				}
				indicationRecieved = true
				if dataComplete && indicationRecieved {
					return nil
				}
			} else {
				return fmt.Errorf("unexpected: %s", data)
			}
		case <-time.After(ub.timeout):
			return ErrTimeout
		}
	}
}

// WaitOnDataChannel waits for data, and calls the passed DataMessageHandler on receipt
// Also handles errors: from the error channel, and timeouts.
func (ub *UbloxBluetooth) WaitOnDataChannel(fn DataMessageHandler) error {
	for {
		select {
		case data := <-ub.DataChannel:
			loop, err := fn(data)
			if !loop {
				return err
			}
		case e := <-ub.ErrorChannel:
			return e
		case <-time.After(ub.timeout):
			return ErrTimeout
		}
	}
}

// HandleDiscovery is used to monitor incoming data channels for discovery
func (ub *UbloxBluetooth) HandleDiscovery(expectedResponse string, fn Discoveryhandler) error {
	var err error
	expected := []byte(expectedResponse)
	loop := true
	for {
		select {
		case data := <-ub.DataChannel:
			if bytes.HasPrefix(data, expected) {
				loop, err = fn(data)
			}
			if err != nil || !loop {
				return err
			}
		case _ = <-ub.CompletedChannel:
			return err
		case e := <-ub.ErrorChannel:
			return e
		case <-time.After(ub.timeout):
			return ErrTimeout
		}
	}
}

func (ub *UbloxBluetooth) processATResponse(b []byte) {
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
