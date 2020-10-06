package ubloxbluetooth

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/8power/ublox-bluetooth/serial"
)

// EMDCmdResp holds EDM CoMmanD and the expected RESPonse
type EMDCmdResp struct {
	Cmd  []byte
	Resp []byte
}

// NewEMDCmdBytes creates an EDM command containing the `payload` content
func NewEMDCmdBytes(payload []byte) []byte {
	l := uint16(len(payload))
	b := make([]byte, l+serial.EDMPayloadOverhead)
	b[0] = serial.EDMStartByte
	binary.BigEndian.PutUint16(b[1:], l)
	copy(b[3:], payload)
	b[3+l] = serial.EDMStopByte
	return b
}

const atPayloadOverhead = 3

// NewEDMATCommand constructs the EDM command from the AT
func NewEDMATCommand(atCommand string) []byte {
	cmd := []byte(atCommand)
	l := len(cmd)
	b := make([]byte, l+atPayloadOverhead)
	b[0] = 0x00
	b[1] = 0x44
	copy(b[2:], cmd)
	b[2+l] = 0x0D
	return NewEMDCmdBytes(b)
}

// ConnectEvent message id
const ConnectEvent = byte(0x11)

// DisconnectEvent message id
const DisconnectEvent = byte(0x21)

// DataEvent message id
const DataEvent = byte(0x31)

// ATRequest message id
const ATRequest = byte(0x44)

// ATConfirmation message id
const ATConfirmation = byte(0x45)

// ATEvent message id
const ATEvent = byte(0x41)

// ResentConnect message id
const ResentConnect = byte(0x56)

// iPhoneEvent message id
const iPhoneEvent = byte(0x61)

// StartEvent message id
const StartEvent = byte(0x71)

func removeNewlines(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte(newline), []byte(""))
}

// ParseEDMMessage parses the message array and extracts the correct message
func (ub *UbloxBluetooth) ParseEDMMessage(msg []byte) error {
	if msg[0] != 0x00 {
		return fmt.Errorf("Message does not start with 0x00")
	}

	data := removeNewlines(msg[2 : len(msg)-1])
	switch msg[1] {
	case StartEvent:
		ub.StartEventReceived = true
	case ATConfirmation:
		switch data[0] {
		case '+':
			ub.DataChannel <- data
		case '"':
			ub.DataChannel <- data[:len(data)-2]
			ub.CompletedChannel <- true
		default:
			ub.handleGeneralMessage(data)
		}
	case ATEvent:
		ub.DataChannel <- data

		// we check for disconnect events disconnectResponse
		if bytes.HasPrefix(data, disconnectResponse) && !ub.disconnectExpected {
			if ub.disconnectHandler != nil {
				go ub.disconnectHandler(ub)
			}
			return ErrUnexpectedDisconnect
		}
	}
	return nil
}
