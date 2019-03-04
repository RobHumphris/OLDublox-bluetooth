package serial

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var verbose = false
var newlineBytes = []byte{'\r', '\n'}

// SetVerbose sets the logging level
func SetVerbose(v bool) {
	verbose = v
}

func showMsg(format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

// SerialPort holds the file and file descriptor for the serial port
type SerialPort struct {
	file             *os.File
	fd               uintptr
	extendedDataMode bool
}

// BaudRate is a type used for enumerating the permissible rates in our system.
type BaudRate uint32

const (
	// Default baud is 115k
	Default BaudRate = unix.B115200
	// HighSpeed baud is 1m
	HighSpeed BaudRate = unix.B1000000
)

// OpenSerialPort opens the specified device with our default settings.
func OpenSerialPort(devicename string, readTimeout time.Duration) (p *SerialPort, err error) {
	f, err := os.OpenFile(devicename, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && f != nil {
			fmt.Printf("ERROR: %v\n", err)
			f.Close()
		}
	}()

	fd := f.Fd()

	unix.SetNonblock(int(fd), false)
	if err != nil {
		return nil, fmt.Errorf("[OpenPort] set non block error: %v", err)
	}

	sp := &SerialPort{
		file: f,
		fd:   fd}

	sp.SetBaudRate(HighSpeed, readTimeout)
	sp.extendedDataMode = false
	return sp, nil
}

// SetEDMFlag is set when we leave AT mode.
func (sp *SerialPort) SetEDMFlag(flag bool) {
	sp.extendedDataMode = flag
}

// SetBaudRate sets the serialport's speed to the passed value
func (sp *SerialPort) SetBaudRate(baudrate BaudRate, readTimeout time.Duration) error {
	br := uint32(baudrate)
	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | unix.IGNCR | br | unix.CS8 | unix.CRTSCTS,
		Ispeed: br,
		Ospeed: br,
	}

	t.Cc[unix.VMIN] = uint8(0x00)
	t.Cc[unix.VTIME] = uint8(readTimeout.Nanoseconds() / 1e6 / 100)

	_, _, errno := unix.Syscall6(
		unix.SYS_IOCTL,
		uintptr(sp.fd),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	)

	if errno != 0 {
		return fmt.Errorf("[OpenPort] ioctl error: %d", errno)
	}
	return nil
}

// Write write's the passed byte array to the serial port
func (sp *SerialPort) Write(b []byte) error {
	showMsg("W: %s\n[%x]", b, b)
	_, err := sp.file.Write(b)
	return err
}

const EDMStartByte = byte(0xAA)
const EDMStopByte = byte(0x55)
const EDMPayloadOverhead = 4
const EDMHeaderSize = 3

// ScanLines reads a complete line from the serial port and sends the bytes
// to the passed channel
func (sp *SerialPort) ScanPort(dataChan chan []byte, edmChan chan []byte, errChan chan error) {
	line := []byte{}
	lineLen := 0
	expectedLength := -1
	edmStartReceived := false
	buf := make([]byte, 1)
	for {
		_, err := sp.file.Read(buf)

		if err != nil {
			if err == io.EOF { // ignore EOFs we're going to get them all the time.
				continue
			} else {
				errChan <- errors.Wrap(err, "serial read error")
			}
		}

		if sp.extendedDataMode {
			if !edmStartReceived {
				if buf[0] == EDMStartByte {
					edmStartReceived = true
				}
			}
			if edmStartReceived {
				line = append(line, buf[0])
				lineLen = len(line)

				if expectedLength == -1 && lineLen == 3 {
					expectedLength = int(binary.BigEndian.Uint16(line[1:3])) + EDMPayloadOverhead
				} else if lineLen == expectedLength {
					if line[expectedLength-1] == EDMStopByte {
						showMsg("EDM R: %s\n[%x]", buf, buf)
						edmChan <- line[EDMHeaderSize:expectedLength]
						line = []byte{}
						expectedLength = -1
						edmStartReceived = false
					} else {
						errChan <- fmt.Errorf("EDM errof Payload length exceeded (Length: %d %x)", expectedLength, line)
						line = []byte{}
						expectedLength = -1
						edmStartReceived = false
					}
				}
			}
		} else {
			line = append(line, buf[0])
			lineLen = len(line)
			if bytes.HasSuffix(line, newlineBytes) {
				if lineLen > 2 {
					showMsg("R: %s\n[%x]", buf, buf)
					dataChan <- line
				}
				line = []byte{}
			}
		}
	}
}

// Ioctl sends
func (sp *SerialPort) ioctl(command int, data int) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(sp.fd),
		uintptr(command),
		uintptr(unsafe.Pointer(&data)),
	)
	if errno != 0 {
		return fmt.Errorf("[Ioctl] error: %d", errno)
	}
	return nil
}

// Flush ensures unwritten bytes are pushed through the serial port.
func (sp *SerialPort) Flush() error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(sp.fd),
		uintptr(0x540B),
		uintptr(unix.TCIOFLUSH),
	)

	if errno != 0 {
		return fmt.Errorf("[Flush] ioctl error: %d", errno)
	}
	return nil
}

// ToggleDTR sets and resets the DTR pin
func (sp *SerialPort) ToggleDTR() error {
	err := sp.ioctl(unix.TIOCMBIS, unix.TIOCM_DTR)
	if err != nil {
		return fmt.Errorf("[ToggleDTR] DTR set error: %d", err)
	}
	time.Sleep(10 * time.Millisecond)

	err = sp.ioctl(unix.TIOCMBIC, unix.TIOCM_DTR)
	if err != nil {
		return fmt.Errorf("[ToggleDTR] DTR clear error: %d", err)
	}
	time.Sleep(10 * time.Millisecond)

	return nil
}

// Close closes the file
func (sp *SerialPort) Close() (err error) {
	return sp.file.Close()
}
