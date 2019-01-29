package ubloxbluetooth

import "fmt"

const discovery = "+UBTD"
const connect = "+UBTACLC"
const connected = "+UUBTACLC"

const errorMessage = "ERROR"
const okMessage = "OK"

func DiscoveryCommand() string {
	return fmt.Sprintf("AT%s", discovery)
}

func ConnectCommand(address string) string {
	return fmt.Sprintf("AT%s=%s", connect, address)
}
