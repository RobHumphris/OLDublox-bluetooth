package serial

import (
	"fmt"
	"testing"

	"github.com/RobHumphris/ublox-bluetooth/serial"
)

func TestGetFTDIDevPath(t *testing.T) {
	path, err := serial.GetFTDIDevPath()
	if err != nil {
		t.Errorf("GetFTDIDevPath failed: %v\n", err)
	}
	fmt.Println("Path returned: ", path)
}
