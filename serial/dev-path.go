package serial

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var PathNotFound = fmt.Errorf("Port not found")
var FormatError = fmt.Errorf("Format Error")

// GetFTDIDevPath gets the current dev path that the USB device is connected to,
// This is based on `/proc/tty/driver/usbserial` containing info in the format of:
// 0: module:ftdi_sio name:"FTDI USB Serial Device" vendor:0403 product:6015 num_ports:1 port:0 path:usb-0000:00:14.0-1.2.2
// or
// 1: module:ftdi_sio name:"FTDI USB Serial Device" vendor:0403 product:6015 num_ports:1 port:0 path:usb-0000:00:14.0-1.2
func GetFTDIDevPath() (string, error) {
	usbserial, err := os.Open("/proc/tty/driver/usbserial")
	if err != nil {
		return "", errors.Wrap(err, "error opening serial driver list")
	}
	defer usbserial.Close()

	scanner := bufio.NewScanner(usbserial)

	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		index := strings.Split(tokens[0], ":")
		module := strings.Split(tokens[1], ":")
		if module[1] == "ftdi_sio" {
			return fmt.Sprintf("/dev/ttyUSB%s", index[0]), nil
		}
	}
	return "", PathNotFound
}
