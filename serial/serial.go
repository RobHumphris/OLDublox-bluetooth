package serial

import (
	"bufio"
	"fmt"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SerialPort holds the file and file descriptor for the SerialPort
type SerialPort struct {
	file *os.File
	fd   uintptr
}

func OpenSerialPort(devicename string, readTimeout time.Duration) (p *SerialPort, err error) {
	baudrate := uint32(0x1002)

	f, err := os.OpenFile(devicename, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && f != nil {
			f.Close()
		}
	}()

	fd := f.Fd()
	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | baudrate | unix.CS8,
		Ispeed: baudrate,
		Ospeed: baudrate,
	}

	t.Cc[unix.VMIN] = uint8(0x00)
	t.Cc[unix.VTIME] = uint8(readTimeout.Nanoseconds() / 1e6 / 100)

	_, _, errno := unix.Syscall6(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	)

	if errno != 0 {
		return nil, fmt.Errorf("[OpenPort] ioctl error: %d", errno)
	}

	unix.SetNonblock(int(fd), false)
	if err != nil {
		return nil, fmt.Errorf("[OpenPort] set non block error: %v", err)
	}

	return &SerialPort{
		file: f,
		fd:   fd}, nil
}

// Write write's the passed byte array to the serial port
func (sp *SerialPort) Write(b []byte) error {
	_, err := sp.file.Write(b)
	return err
}

// ScanLines reads a complete line from the serial port and sends the bytes
// to the passed channel
func (sp *SerialPort) ScanLines(ch chan []byte) {
	fmt.Println("[ScanLines] Start")
	defer fmt.Println("[ScanLines] End")
	for {
		scanner := bufio.NewScanner(sp.file)
		for scanner.Scan() {
			ch <- scanner.Bytes()
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
