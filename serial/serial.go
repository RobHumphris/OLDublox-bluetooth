package serial

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

var verbose = false

// SetVerbose sets the logging level
func SetVerbose(v bool) {
	verbose = v
}

func showMsg(format string, v ...interface{}) {
	if verbose {
		fmt.Printf(format, v...)
	}
}

// SerialPort holds the file and file descriptor for the serial port
type SerialPort struct {
	file *os.File
	fd   uintptr
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

	return sp, nil
}

// SetBaudRate sets the serialport's speed to the passed value
func (sp *SerialPort) SetBaudRate(baudrate BaudRate, readTimeout time.Duration) error {
	br := uint32(baudrate)
	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | unix.IGNCR | br | unix.CS8,
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
	showMsg("W: %s", b)
	_, err := sp.file.Write(b)
	return err
}

// Read a single byte from the SerialPort
func (sp *SerialPort) Read() ([]byte, error) {
	a := []byte{0}
	_, err := sp.file.Read(a)
	return a, err
}

// ReadLine reads a complete line from the serialPort
func (sp *SerialPort) ReadLine() ([]byte, error) {
	scanner := bufio.NewScanner(sp.file)
	for scanner.Scan() {
		return scanner.Bytes(), scanner.Err()
	}
	return nil, nil
}

var newline = []byte{'\r', '\n'}

// ScanLines reads a complete line from the serial port and sends the bytes
// to the passed channel
func (sp *SerialPort) ScanLines(ch chan []byte) {
	line := []byte{}
	buf := make([]byte, 1)
	for {
		_, err := sp.file.Read(buf)
		if err != nil {
		}
		line = append(line, buf[0])
		if bytes.HasSuffix(line, newline) {
			if len(line) > 2 {
				showMsg("R: %s", line)
				ch <- line
			}
			line = []byte{}
		}
	}
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

// Close closes the file
func (sp *SerialPort) Close() (err error) {
	return sp.file.Close()
}
