package ubloxbluetooth

type DiscoveryReply struct {
	BluetoothAddress string
	Rssi             int
	DeviceName       string
	DataType         int
	Data             string
}

type ConnectionReply struct {
	Handle           int
	Type             int
	BluetoothAddress string
}

type VersionReply struct {
	SoftwareVersion int
	HardwareVersion int
}

type InfoReply struct {
	CurrentTime           int
	CurrentSequenceNumber int
	RecordsCount          int
}

type ConfigReply struct {
	AdvertisingInterval int
	SampleTime          int
	State               int
	AccelSettings       int
	SpareOne            int
	TemperatureOffset   int
}

type SlotCountReply struct {
	Count    int
	rawCount string
}

type SlotInfoReply struct {
	Time int
	t    int
}
