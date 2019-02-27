package ubloxbluetooth

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

// ConnectDeviceSPS enables serial port service (data pump) on the device
func (ub *UbloxBluetooth) ConnectDeviceSPS(macAddress string) (int, error) {
	url := fmt.Sprintf("sps://%s", macAddress)
	b, err := ub.writeAndWait(ConnectPeerCommand(url), true)
	if err != nil {
		return -1, errors.Wrap(err, "ConnectDeviceSPS error")
	}

	handle, err := ProcessConnectDeviceSPSReply(string(b))
	if err != nil {
		return -1, errors.Wrap(err, "ConnectDeviceSPS error")
	}

	var peer *ConnectedPeer
	var acl *ACLConnected
	err = ub.WaitOnDataChannel(func(data []byte) (bool, error) {
		if bytes.HasPrefix(data, peerConnectedResponse) {
			p, err := NewConnectedPeerReply(string(data))
			if err != nil {
				return false, errors.Wrap(err, "NewConnectedPeerReply error")
			}
			peer = p
			return false, nil
		} else if bytes.HasPrefix(data, aclConnectionRemoteDeviceResponse) {
			a, err := NewACLConnectedReply(string(data))
			if err != nil {
				return false, errors.Wrap(err, "NewACLConnectedReply error")
			}
			acl = a
			return true, nil
		}
		return true, nil
	})
	if err != nil {
		return handle, errors.Wrap(err, "WaitOnDataChannel error")
	}

	if acl != nil {
		if acl.MacAddress != macAddress {
			return handle, fmt.Errorf("NewConnectedPeerReply ACL Mac Addresses do not match (wanted: %s got: %s)", macAddress, acl.MacAddress)
		}
	}

	if peer.PeerHandle != handle {
		return handle, fmt.Errorf("NewConnectedPeerReply error handles do not match (wanted: %d got: %d)", handle, peer.PeerHandle)
	}

	return handle, err
}

// DisconnectDeviceSPS disconnects from the device with the given peerHandle
func (ub *UbloxBluetooth) DisconnectDeviceSPS(peerHandle int) error {
	err := ub.EnterCommandMode()
	if err != nil {
		return errors.Wrap(err, "EnterCommandMode error")
	}

	d, err := ub.writeAndWait(DisconnectPeerCommand(peerHandle), true)
	if err != nil {
		return errors.Wrap(err, "DisconnectPeerCommand error")
	}

	return ProcessPeerDisconnectedReply(peerHandle, string(d))
}

// WriteSPS writes the bytes to the serial port service
func (ub *UbloxBluetooth) WriteSPS(d []byte) error {
	if ub.currentMode != dataMode {
		return fmt.Errorf("WriteSPS error. Not in Data Mode")
	}
	return ub.WriteBytes(d)
}
