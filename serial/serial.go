package serial

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

var newlineBytes = []byte{'\r', '\n'}

const (
	// EDMStartByte Extended Data Mode start byte value
	EDMStartByte = byte(0xAA)
	// EDMStopByte Extended Data Mode stop byte value
	EDMStopByte = byte(0x55)
	// EDMPayloadOverhead the number of bytes to skip to the start of the payload
	EDMPayloadOverhead = 4
	// EDMHeaderSize Extended Data Mode header size
	EDMHeaderSize = 3
)

// SetVerbose sets the logging level
func (sp *SerialPort) SetVerbose(v bool) {
	sp.verbose = v
}

func (sp *SerialPort) showOutMsg(b []byte) {
	if sp.verbose {
		l := len(b) - 2
		fmt.Printf("-> %s\n", b[5:l])
	}
}

// GetPortStats gets the comms stats for this port
func (sp *SerialPort) GetPortStats() *SerialPortStats {
	// return a copy
	rs := &SerialPortStats{
		TxBytes: sp.stats.TxBytes,
		RxBytes: sp.stats.RxBytes,
	}
	return rs
}

func (sp *SerialPort) showInMsg(b []byte) {
	if sp.verbose {
		l := len(b) - 3
		if l > 7 {
			fmt.Printf("<- %s\n", b[7:l])
		} else {
			fmt.Printf("<- %+q\n", b)
		}
	}
}

// SerialPortStats hold useful stats for debugging
type SerialPortStats struct {
	TxBytes uint64
	RxBytes uint64
}

// SerialPort holds the file and file descriptor for the serial port
type SerialPort struct {
	fd               int
	extendedDataMode bool
	contineScanning  bool
	isOpen           bool
	byteBuf          []byte
	verbose          bool
	stats            *SerialPortStats
}

// BaudRate is a type used for enumerating the permissible rates in our system.
type BaudRate uint32

const (
	// Default baud is 115k
	Default BaudRate = unix.B115200
	// HighSpeed baud is 1m
	HighSpeed BaudRate = unix.B1000000
)

// OpenSerialPort opens a Ublox device with a timeout value
func OpenSerialPort(devPath string, readTimeout time.Duration) (p *SerialPort, err error) {
	fd, err := syscall.Open(devPath, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			fmt.Printf("[OpenSerialPort] ERROR: %v\n", err)
			syscall.Close(fd)
		}
	}()

	syscall.SetNonblock(fd, false)
	if err != nil {
		return nil, fmt.Errorf("[OpenSerialPort] set non block error: %v", err)
	}

	sp := &SerialPort{
		fd:               fd,
		extendedDataMode: true,
		contineScanning:  true,
		isOpen:           true,
		byteBuf:          make([]byte, 1),
		verbose:          false,
		stats: &SerialPortStats{
			TxBytes: 0,
			RxBytes: 0,
		},
	}

	sp.SetBaudRate(HighSpeed, readTimeout)
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
	sp.showOutMsg(b)
	n, err := syscall.Write(sp.fd, b)
	sp.stats.TxBytes += uint64(n)
	return err
}

func (sp *SerialPort) read(readChan chan byte) error {
	var err error
	buff := make([]byte, 1)
	for sp.contineScanning {
		n, err := syscall.Read(sp.fd, buff)
		if err != nil {
			sp.contineScanning = false
			break
		}
		if n > 0 {
			sp.stats.RxBytes++
			readChan <- buff[0]
		}
	}
	return err
}

// StopScanning sets the continueScanning flag to false
func (sp *SerialPort) StopScanning() {
	sp.contineScanning = false
}

// ScanPort reads a complete line from the serial port and sends the bytes
// to the passed channel
func (sp *SerialPort) ScanPort(ctx context.Context, dataChan chan []byte, edmChan chan []byte, errChan chan error) error {
	var err error
	fmt.Println("[ScanPort] starting")

	line := []byte{}
	lineLen := 0
	expectedLength := -1
	edmStartReceived := false
	sp.contineScanning = true

	rchan := make(chan byte, 1)
	go sp.read(rchan)

	for sp.contineScanning {
		select {
		case <-ctx.Done():
			fmt.Printf("[ScanPort] context done")
			return err
		case b := <-rchan:
			if sp.extendedDataMode {
				if !edmStartReceived {
					if b == EDMStartByte {
						edmStartReceived = true
					}
				}
				if edmStartReceived {
					line = append(line, b)
					lineLen = len(line)

					if expectedLength == -1 && lineLen == 3 {
						expectedLength = int(binary.BigEndian.Uint16(line[1:3])) + EDMPayloadOverhead
					} else if lineLen == expectedLength {
						if line[expectedLength-1] == EDMStopByte {
							sp.showInMsg(line)
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
				line = append(line, b)
				lineLen = len(line)
				if bytes.HasSuffix(line, newlineBytes) {
					if lineLen > 2 {
						sp.showInMsg(line)
						dataChan <- line
					}
					line = []byte{}
				}
			}

		}

		/*b, err := sp.read()
		if err != nil {
			if err == io.EOF { // ignore EOFs we're going to get them all the time.
				continue
			} else {
				if sp.isOpen {
					errChan <- errors.Wrap(err, "serial read error")
				} else {
					fmt.Printf("[ScanPort] Read error %v\n", err)
				}
				break
			}
		}

		if sp.extendedDataMode {
			if !edmStartReceived {
				if b == EDMStartByte {
					edmStartReceived = true
				}
			}
			if edmStartReceived {
				line = append(line, b)
				lineLen = len(line)

				if expectedLength == -1 && lineLen == 3 {
					expectedLength = int(binary.BigEndian.Uint16(line[1:3])) + EDMPayloadOverhead
				} else if lineLen == expectedLength {
					if line[expectedLength-1] == EDMStopByte {
						sp.showInMsg(line)
						edmChan <- line[EDMHeaderSize:expectedLength]
						line = []byte{}
						expectedLength = -1ctx
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
			line = append(line, b)
			lineLen = len(line)
			if bytes.HasSuffix(line, newlineBytes) {
				if lineLen > 2 {
					sp.showInMsg(line)
					dataChan <- line
				}
				line = []byte{}
			}
		}*/
	}
	return err
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

var defaultDTRPause = 10 * time.Millisecond

func (sp *SerialPort) setDTR() error {
	err := sp.ioctl(unix.TIOCMBIS, unix.TIOCM_DTR)
	if err != nil {
		return fmt.Errorf("[ToggleDTR] DTR set error: %d", err)
	}
	time.Sleep(defaultDTRPause)
	return nil
}

func (sp *SerialPort) clearDTR() error {
	err := sp.ioctl(unix.TIOCMBIC, unix.TIOCM_DTR)
	if err != nil {
		return fmt.Errorf("[ToggleDTR] DTR clear error: %d", err)
	}
	time.Sleep(defaultDTRPause)
	return nil
}

// ResetViaDTR sends the DTR line low and then takes it high
// if the board has been setup with AT&D4 this will cause a reset.
func (sp *SerialPort) ResetViaDTR() error {
	err := sp.clearDTR()
	if err != nil {
		return err
	}

	err = sp.setDTR()
	if err != nil {
		return err
	}

	return nil
}

// Close closes the file
func (sp *SerialPort) Close() (err error) {
	err = syscall.Close(sp.fd)
	sp.isOpen = false
	return err
}
