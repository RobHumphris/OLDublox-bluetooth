package ubloxbluetooth

// DiscoveryReply BLE discovery structure
type DiscoveryReply struct {
	BluetoothAddress string
	Rssi             int
	DeviceName       string
	DataType         int
	Data             string
}

// RS232SettingsReply serial port settings structure
type RS232SettingsReply struct {
	BaudRate           int
	FlowControl        int
	DataBits           int
	StopBits           int
	Parity             int
	ChangeAfterConfirm int
}

// ConnectionReply connection data structure
type ConnectionReply struct {
	Handle           int
	Type             int
	BluetoothAddress string
}

// VersionReply VEH sensor version structure
type VersionReply struct {
	SoftwareVersion int
	HardwareVersion int
}

// InfoReply sensor info (time, sequence, & count) structure
type InfoReply struct {
	CurrentTime           int
	CurrentSequenceNumber int
	RecordsCount          int
}

// ConfigReply sensor conf structure
type ConfigReply struct {
	AdvertisingInterval int
	SampleTime          int
	State               int
	AccelSettings       int
	SpareOne            int
	TemperatureOffset   int
}

// SlotCountReply holds the slot data
type SlotCountReply struct {
	Count    int
	rawCount string
}

// SlotInfoReply holds the current data returned for Info
type SlotInfoReply struct {
	Time           int
	Slot           int
	Bytes          int
	SampleRate     float32
	Temperature    int
	BatteryVoltage int
	VoltageIn      int
}

// ConnectedPeer describes the Bluetooth peer's connection
type ConnectedPeer struct {
	PeerHandle int
	Type       int
	Profile    int
	MacAddress string
	FrameSize  int
}

type ACLConnected struct {
	ConnHandle int
	Type       int
	MacAddress string
}
