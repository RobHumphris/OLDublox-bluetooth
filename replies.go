package ubloxbluetooth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type DiscoveryReply struct {
	BluetoothAddress string
	Rssi             int
	DeviceName       string
	DataType         int
	Data             string
}

// NewDiscoveryReply takes the string format:
// p: 506B389221F8r,-76,"",2,02010607FF4C0010020B08
// and converts it to a DiscoveryReply
func NewDiscoveryReply(d string) (*DiscoveryReply, error) {
	t := strings.Split(d, ",")
	if len(t) < 5 {
		return nil, fmt.Errorf("[NewDiscoveryReply] Not enough tokens in string")
	}

	rssi, err := strconv.Atoi(t[1])
	if err != nil {
		return nil, errors.Wrap(err, "[NewDiscoveryReply] error extracting RSSI")
	}

	dataType, err := strconv.Atoi(t[3])
	if err != nil {
		return nil, errors.Wrap(err, "[NewDiscoveryReply] error extracting DataType")
	}

	return &DiscoveryReply{
		BluetoothAddress: t[0],
		Rssi:             rssi,
		DeviceName:       t[2],
		DataType:         dataType,
		Data:             t[4],
	}, nil
}
