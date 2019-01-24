package ubloxbluetooth

import (
	"fmt"
	"strings"
	"time"

	"github.com/RobHumphris/ublox-bluetooth/serial"
)

// UbloxBluetooth holds the serial port, and the communication channels.
type UbloxBluetooth struct {
	serialPort       *serial.SerialPort
	readChannel      chan []byte
	discoveryChannel chan *DiscoveryReply
	errorChannel     chan error
}

// NewUbloxBluetooth creates a new UbloxBluetooth instance
func NewUbloxBluetooth(device string, timeout time.Duration) (*UbloxBluetooth, error) {
	sp, err := serial.OpenSerialPort(device, timeout)
	if err != nil {
		return nil, err
	}

	ub := &UbloxBluetooth{
		serialPort:       sp,
		readChannel:      make(chan []byte),
		discoveryChannel: make(chan *DiscoveryReply),
		errorChannel:     make(chan error),
	}

	go ub.serialportReader()

	return ub, nil
}

var top = []byte("AT")
var tail = []byte("\r\n")

// Write writes the data string to Ublox via the SerialPort
func (ub *UbloxBluetooth) Write(data string) error {
	fmt.Printf("Writing %s to u-blox\n", data)
	d := append(top, []byte(data)...)
	return ub.serialPort.Write([]byte(append(d, tail...)))
}

func (ub *UbloxBluetooth) serialportReader() {
	go ub.serialPort.ScanLines(ub.readChannel)
	for {
		b := <-ub.readChannel
		if len(b) != 0 {
			str := string(b[:])
			tok := strings.Split(str, ":")
			if len(tok) > 1 {
				switch tok[0] {
				case Discovery:
					ub.handleScanPayload(tok[1])
				default:
					ub.handleUnknownPayload(tok[1])
				}
			} else {
				ub.handleUnknownPayload(tok[0])
			}
		}
	}
}

func (ub *UbloxBluetooth) handleScanPayload(p string) {
	discovered, err := NewDiscoveryReply(p)
	if err != nil {
		ub.errorChannel <- err
	} else {
		ub.discoveryChannel <- discovered
	}
}

func (ub *UbloxBluetooth) handleUnknownPayload(p string) {
	ub.errorChannel <- fmt.Errorf("Unknown payload %s", p)
}
