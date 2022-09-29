package ubloxbluetooth

import (
	"encoding/binary"
	"fmt"
	"math"
)

// VEH Sensor event types
const (
	VehEventBoot         = 0x00
	VehEventSystemOff    = 0x01
	VehEventTimeAdjust   = 0x02
	VehEventLogCleared   = 0x05
	VehEventMessage      = 0x06
	VehEventEventLog     = 0x0A // Not used
	VehEventDiectory     = 0x0B // Can be ignored
	VehEventWatchdog     = 0x0C
	VehEventAppErr       = 0x0D
	VehEventAssert       = 0x0E
	VehEventHardFault    = 0x0F
	VehEventEfmStatus    = 0x10
	VehEventDummy        = 0x16
	VehEventTemperature  = 0x64
	VehEventVibration    = 0x65
	VehEventMicrophone   = 0x66
	VehEventHallEffect   = 0x67
	VehEventConnected    = 0xC0
	VehEventDisconnected = 0xD0
	VehEventAlert        = 0x0E
	VehEventUnused       = 0xFC
	VehEventBad          = 0xFD
	VehEventDataLoss     = 0xFE
	VehEventError        = 0xFF
)

type field struct {
	offset int
	size   int
}

func (f *field) Start() int {
	return f.offset
}

func (f *field) End() int {
	return f.offset + f.size
}

const (
	sizeofFloat32 = 4
	sizeofUint32  = 4
	sizeofUint16  = 2
	sizeofUint8   = 1
	variableSize  = 0
)

var (
	/*
		VEH_EVT_HDR
		Offset		Parameter		Type			Description
		0			seqno			uint32_t		Sequence number
		4			timestamp		uint32_t		Timestamp in seconds
		8			tag				uint8_t			Event type
		9			length			uint8_t			Length of payload (bytes)
		10			Payload			uint8_t[]		Event payload
		10+length	flag			uint8_t			Flag to indicate presence of data
	*/
	ehSeqNoOffset     field = field{0, sizeofUint32}
	ehTimestampOffset field = field{4, sizeofUint32}
	ehEventTypeOffset field = field{8, sizeofUint8}
	ehLengthOffset    field = field{9, sizeofUint8}
	ehPayloadOffset   field = field{10, variableSize}

	/*
		VEH_EVT_BOOT 0x00
		Offset		Parameter		Type			Description
		10			reason			uint32_t		CPU reset register contents
		14			version			uint8[4]		Version
		18			flag			uint8_t			0 = no data
	*/
	ebReasonOffset  field = field{10, sizeofUint32}
	ebVersionOffset field = field{14, sizeofUint32}
	ebFlagOffset    field = field{18, sizeofUint8}

	/*
		VEH_EVT_SYSTEM_OFF 0x01
		Offset		Parameter		Type			Description
		10			flag			uint8_t			0 = no data
	*/
	esoFlagOffset field = field{10, sizeofUint8}

	/*
		VEH_EVT_TIME_ADJUST 0x02
		Offset		Parameter		Type			Description
		10			current			uint32_t		Current time value
		14			updated			uint32_t		Updated time value
		18			flag			uint8_t			0 = no data
	*/
	etaCurrentOffset field = field{10, sizeofUint32}
	etaUpdatedOffset field = field{14, sizeofUint32}
	etaFlagOffset    field = field{18, sizeofUint8}

	/*
		VEH_EVT_LOG_CLEARED 0x05
		Offset		Parameter		Type			Description
		10			flag			uint8_t			0 = no data
	*/
	elcFlagOffset field = field{10, sizeofUint8}

	/*
		VEH_EVT_MESSAGE 0x06
		Offset		Parameter		Type			Description
		10			message			uint8_t[]		Message writtn to event log
		10+length	flag			uint8_t			0 = no data
	*/
	emMessageOffset field = field{10, variableSize}

	/*
		VEH_EVT_WATCHDOG 0x0C
		Offset		Parameter		Type			Description
		10			r0				uint32_t		R0 register
		14			r1				uint32_t		R1 register
		18			r2				uint32_t		R2 register
		22			r3				uint32_t		R3 register
		26			r12				uint32_t		R12 register
		30			lr				uint32_t		Link register
		34			pc				uint32_t		Program counter
		38			psr				uint32_t		Pragram Status register
	*/
	ewdR0Offset  field = field{10, sizeofUint32}
	ewdR1Offset  field = field{14, sizeofUint32}
	ewdR2Offset  field = field{18, sizeofUint32}
	ewdR3Offset  field = field{22, sizeofUint32}
	ewdR12Offset field = field{26, sizeofUint32}
	ewdLROffset  field = field{30, sizeofUint32}
	ewdPCOffset  field = field{34, sizeofUint32}
	ewdPSROffset field = field{38, sizeofUint32}

	/*
		VEH_EVT_APP_ERR 0x0D
		Offset		Parameter		Type			Description
		10			line_num		uint16_t		the line number where the error occurred
		12			file_name		uint8_t[]		the file name in which the error occurred
		12+length	err_code		uint32_t		Error code
	*/
	eaeLineNoOffset   field = field{10, sizeofUint16}
	eaeFileNameOffset field = field{12, variableSize}

	/*
		VEH_EVT_ASSERT 0x0E
		Offset		Parameter		Type			Description
		10			line_num		uint16_t		the line number where the assert occurred
		12			file_name		uint8_t[]		the file name in which the assert occurred
	*/
	easLineNoOffset   field = field{10, sizeofUint16}
	easFileNameOffset field = field{12, variableSize}

	/*
		VEH_EVT_HARDFAULT 0x0F
		Offset		Parameter		Type			Description
		10			r0				uint32_t		R0 register
		14			r1				uint32_t		R1 register
		18			r2				uint32_t		R2 register
		22			r3				uint32_t		R3 register
		26			r12				uint32_t		R12 register
		30			lr				uint32_t		Link register
		34			pc				uint32_t		Program counter
		38			psr				uint32_t		Pragram Status register
	*/
	ehfR0Offset  field = field{10, sizeofUint32}
	ehfR1Offset  field = field{14, sizeofUint32}
	ehfR2Offset  field = field{18, sizeofUint32}
	ehfR3Offset  field = field{22, sizeofUint32}
	ehfR12Offset field = field{26, sizeofUint32}
	ehfLROffset  field = field{30, sizeofUint32}
	ehfPCOffset  field = field{34, sizeofUint32}
	ehfPSROffset field = field{38, sizeofUint32}

	/*
		VEH_EVT_APP_TEMPERATURE 0x64
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			humidity		float32_t		Relative humidity (%)
		22			bat%			float32_t		Battery charge as a percentage (100%)
		26			flag			uint8_t			0 = no data
	*/
	etBatteryOffset     field = field{10, sizeofFloat32}
	etTemperatureOffset field = field{14, sizeofFloat32}
	etHumidityOffset    field = field{18, sizeofFloat32}
	etBatteryPercOffset field = field{22, sizeofFloat32}
	etFlagOffset        field = field{26, sizeofUint8}

	/*
		VEH_EVT_APP_VIBRATION 0x65
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			odr				float32_t		Estimated sample rate freq (Hz)
		22			gain			float32_t		Gain setting
		26			bat%			float32_t		Battery charge as a percentage (100%)
		30			flag			uint8_t			0 = no data, 1 = data
	*/
	evBatteryOffset     field = field{10, sizeofFloat32}
	evTemperatureOffset field = field{14, sizeofFloat32}
	evOdrOffset         field = field{18, sizeofFloat32}
	evGainOffset        field = field{22, sizeofFloat32}
	evBatteryPercOffset field = field{26, sizeofFloat32}
	evFlagOffset        field = field{30, sizeofUint8}

	/*
		VEH_EVT_APP_MICROPHONE 0x66
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			odr				float32_t		Output data rate
		22			battery%		float32_t		Battery gauge as a percentage
		26			flag			uint8_t			0 = no data
	*/
	emBatteryOffset     field = field{10, sizeofFloat32}
	emTemperatureOffset field = field{14, sizeofFloat32}
	emOdrOffset         field = field{18, sizeofFloat32}
	emBatteryPercOffset field = field{22, sizeofFloat32}
	emFlagOffset        field = field{26, sizeofUint8}

	/*
		VEH_EVT_APP_HALL 0x67
		Offset		Parameter		Type			Description
		10			battery			float32_t		Battery voltage (Volts)
		14			temperature		float32_t		Current temperature (C)
		18			odr				float32_t		Output data rate
		22			hysteresis		float32_t		0 = no hysteresis
		26			range			float32_t		Range in tesla
		30			fir_filter		uint16_t		Finite Impulse response filter burst size
		32			battery%		float32_t		Battery gauge as a percentage
		36			flag			uint8_t			0 = no data
	*/
	ehBatteryOffset     field = field{10, sizeofFloat32}
	ehTemperatureOffset field = field{14, sizeofFloat32}
	ehOdrOffset         field = field{18, sizeofFloat32}
	ehHysteresisOffset  field = field{22, sizeofFloat32}
	ehRangeOffset       field = field{26, sizeofFloat32}
	ehFirFilterOffset   field = field{30, sizeofUint16}
	ehBatteryPercOffset field = field{32, sizeofFloat32}
	ehFlagOffset        field = field{36, sizeofUint8} // Currently too long, needs sorting

	/*
		VEH_EVT_ERROR 0x05
		Offset		Parameter		Type			Description
		10			flag			uint8_t			0 = no data
	*/
	eerFlagOffset field = field{10, sizeofUint8}
)

type handler func([]byte) *VehEvent
type handlersMap map[byte]handler

var handlers handlersMap

// VehEvent is the base interface
type VehEvent struct {
	DataFlag          bool
	Sequence          uint32
	Timestamp         uint32
	EventType         int
	BootEvent         *VehBootEvent
	TimeAdjustEvent   *VehTimeAdjustEvent
	MessageEvent      *VehMessageEvent
	WatchDogEvent     *VehWatchDogEvent
	AppErrEvent       *VehAppErrEvent
	AssertEvent       *VehAssertEvent
	HardFaultEvent    *VehHardFaultEvent
	TemperatureEvent  *VehTemperatureEvent
	VibrationEvent    *VehVibrationEvent
	MicrophoneEvent   *VehMicrophoneEvent
	HallEvent         *VehHallEvent
	ConnectedEvent    *VehConnectedEvent
	DisconnectedEvent *VehDisconnectedEvent
}

// VehBootEvent structure
type VehBootEvent struct {
	Reason          uint32
	SoftwareVersion string
	HardwareVersion uint32
	BuildNumber     uint32
}

type VehTimeAdjustEvent struct {
	Current uint32
	Updated uint32
}

type VehMessageEvent struct {
	Message string
}

type VehWatchDogEvent struct {
	R0  uint32
	R1  uint32
	R2  uint32
	R3  uint32
	R12 uint32
	LR  uint32
	PC  uint32
	PSR uint32
}

type VehAppErrEvent struct {
	LineNo    uint32
	Filename  string
	ErrorCode uint32
}

type VehAssertEvent struct {
	LineNo   uint32
	Filename string
}

type VehHardFaultEvent struct {
	R0  uint32
	R1  uint32
	R2  uint32
	R3  uint32
	R12 uint32
	LR  uint32
	PC  uint32
	PSR uint32
}

// VehConnectedEvent structure
type VehConnectedEvent struct {
	Mac string
}

// VehDisconnectedEvent structure
type VehDisconnectedEvent struct {
	Mac string
}

// VehTemperatureEvent structure
type VehTemperatureEvent struct {
	Battery     float32
	Temperature float32
	Humidity    float32
	BatPercent  float32
}

// VehVibrationEvent structure
type VehVibrationEvent struct {
	Battery     float32
	Temperature float32
	Odr         float32
	Gain        float32
	BatPercent  float32
}

type VehMicrophoneEvent struct {
	Battery     float32
	Temperature float32
	Odr         float32
	BatPercent  float32
}

type VehHallEvent struct {
	Battery     float32
	Temperature float32
	Odr         float32
	Hysteresis  float32
	Range       float32
	FirFilter   uint16
	BatPercent  float32
}

func init() {
	handlers = make(handlersMap)
	handlers[VehEventBoot] = newBootEvent
	handlers[VehEventSystemOff] = newGenericEvent
	handlers[VehEventTimeAdjust] = newTimeAdjustEvent
	handlers[VehEventLogCleared] = newGenericEvent
	handlers[VehEventMessage] = newMessageEvent
	handlers[VehEventWatchdog] = newWatchDogEvent
	handlers[VehEventAppErr] = newAppErrEvent
	handlers[VehEventAssert] = newAssertEvent
	handlers[VehEventHardFault] = newHardFaultEvent
	handlers[VehEventTemperature] = newTemperatureEvent
	handlers[VehEventVibration] = newVibrationEvent
	handlers[VehEventMicrophone] = newMicrophoneEvent
	handlers[VehEventHallEffect] = newHallEvent
	handlers[VehEventConnected] = newConnectedEvent
	handlers[VehEventDisconnected] = newDisconnectedEvent
	handlers[VehEventError] = newGenericEvent
}

// NewRecorderEvent returns new RecorderEvent
func NewRecorderEvent(b []byte) (*VehEvent, error) {
	if len(b) < ehLengthOffset.Start() {
		return nil, ErrorTruncatedResponse
	}

	fn, ok := handlers[b[ehEventTypeOffset.Start()]]
	if ok {
		return fn(b), nil
	}
	return nil, fmt.Errorf("unhandled Event type: %02X", b[ehEventTypeOffset.Start()])
}

func newGenericEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	return &eb
}

func newVehEvent(b []byte) (VehEvent, int) {
	length := 10 + int(b[ehLengthOffset.Start()])
	return VehEvent{
		DataFlag:  b[length] > 0x00,
		Sequence:  binary.LittleEndian.Uint32(b[ehSeqNoOffset.Start():ehSeqNoOffset.End()]),
		Timestamp: binary.LittleEndian.Uint32(b[ehTimestampOffset.Start():ehTimestampOffset.End()]),
		EventType: int(b[ehEventTypeOffset.Start()]),
	}, length
}

func newBootEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.BootEvent = &VehBootEvent{
		binary.LittleEndian.Uint32(b[ebReasonOffset.Start():ebReasonOffset.End()]),
		fmt.Sprintf("%d.%d", b[ebVersionOffset.Start()], b[ebVersionOffset.Start()+1]),
		uint32(b[ebVersionOffset.Start()+2]),
		uint32(b[ebVersionOffset.Start()+3]),
	}
	return &eb
}

func newTimeAdjustEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.TimeAdjustEvent = &VehTimeAdjustEvent{
		Current: binary.LittleEndian.Uint32(b[etaCurrentOffset.Start():etaCurrentOffset.End()]),
		Updated: binary.LittleEndian.Uint32(b[etaUpdatedOffset.Start():etaUpdatedOffset.End()]),
	}
	return &eb
}

func newMessageEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.MessageEvent = &VehMessageEvent{
		Message: string(b[emMessageOffset.Start() : emMessageOffset.Start()+l]),
	}
	return &eb
}

func newWatchDogEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.WatchDogEvent = &VehWatchDogEvent{
		R0:  binary.LittleEndian.Uint32(b[ewdR0Offset.Start():ewdR0Offset.End()]),
		R1:  binary.LittleEndian.Uint32(b[ewdR1Offset.Start():ewdR1Offset.End()]),
		R2:  binary.LittleEndian.Uint32(b[ewdR2Offset.Start():ewdR2Offset.End()]),
		R3:  binary.LittleEndian.Uint32(b[ewdR3Offset.Start():ewdR3Offset.End()]),
		R12: binary.LittleEndian.Uint32(b[ewdR12Offset.Start():ewdR12Offset.End()]),
		LR:  binary.LittleEndian.Uint32(b[ewdLROffset.Start():ewdLROffset.End()]),
		PC:  binary.LittleEndian.Uint32(b[ewdPCOffset.Start():ewdPCOffset.End()]),
		PSR: binary.LittleEndian.Uint32(b[ewdPSROffset.Start():ewdPSROffset.End()]),
	}
	return &eb
}

func newAppErrEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.AppErrEvent = &VehAppErrEvent{
		LineNo:    binary.LittleEndian.Uint32(b[eaeLineNoOffset.Start():eaeLineNoOffset.End()]),
		Filename:  string(b[eaeFileNameOffset.Start() : eaeFileNameOffset.Start()+l-sizeofUint32]),
		ErrorCode: binary.LittleEndian.Uint32(b[eaeFileNameOffset.Start()+l-sizeofUint32 : eaeFileNameOffset.Start()+l]),
	}
	return &eb
}

func newAssertEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.AssertEvent = &VehAssertEvent{
		LineNo:   binary.LittleEndian.Uint32(b[easLineNoOffset.Start():easLineNoOffset.End()]),
		Filename: string(b[easFileNameOffset.Start() : easFileNameOffset.Start()+l]),
	}
	return &eb
}

func newHardFaultEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.HardFaultEvent = &VehHardFaultEvent{
		R0:  binary.LittleEndian.Uint32(b[ehfR0Offset.Start():ehfR0Offset.End()]),
		R1:  binary.LittleEndian.Uint32(b[ehfR1Offset.Start():ehfR1Offset.End()]),
		R2:  binary.LittleEndian.Uint32(b[ehfR2Offset.Start():ehfR2Offset.End()]),
		R3:  binary.LittleEndian.Uint32(b[ehfR3Offset.Start():ehfR3Offset.End()]),
		R12: binary.LittleEndian.Uint32(b[ehfR12Offset.Start():ehfR12Offset.End()]),
		LR:  binary.LittleEndian.Uint32(b[ehfLROffset.Start():ehfLROffset.End()]),
		PC:  binary.LittleEndian.Uint32(b[ehfPCOffset.Start():ehfPCOffset.End()]),
		PSR: binary.LittleEndian.Uint32(b[ehfPSROffset.Start():ehfPSROffset.End()]),
	}
	return &eb
}

func newTemperatureEvent(b []byte) *VehEvent {
	eb, _ := newVehEvent(b)
	eb.TemperatureEvent = &VehTemperatureEvent{
		Battery:     float32FromBytes(b[etBatteryOffset.Start():etBatteryOffset.End()]),
		Temperature: float32FromBytes(b[etTemperatureOffset.Start():etTemperatureOffset.End()]),
		Humidity:    float32FromBytes(b[etHumidityOffset.Start():etHumidityOffset.End()]),
		BatPercent:  float32FromBytes(b[etBatteryPercOffset.Start():etBatteryPercOffset.End()]),
	}
	return &eb
}

func newVibrationEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.VibrationEvent = &VehVibrationEvent{
		Battery:     float32FromBytes(b[evBatteryOffset.Start():evBatteryOffset.End()]),
		Temperature: float32FromBytes(b[evTemperatureOffset.Start():evTemperatureOffset.End()]),
		Odr:         0.0,
		Gain:        0.0,
		BatPercent:  0.0,
	}
	if l >= evOdrOffset.End() {
		eb.VibrationEvent.Odr = float32FromBytes(b[evOdrOffset.Start():evOdrOffset.End()])
	}
	if l >= evGainOffset.End() {
		eb.VibrationEvent.Gain = float32FromBytes(b[evGainOffset.Start():evGainOffset.End()])
	}
	if l >= evBatteryPercOffset.End() {
		eb.VibrationEvent.BatPercent = float32FromBytes(b[evBatteryPercOffset.Start():evBatteryPercOffset.End()])
	}
	return &eb
}

func newMicrophoneEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.MicrophoneEvent = &VehMicrophoneEvent{
		Battery:     float32FromBytes(b[emBatteryOffset.Start():emBatteryOffset.End()]),
		Temperature: float32FromBytes(b[emTemperatureOffset.Start():emTemperatureOffset.End()]),
		Odr:         float32FromBytes(b[emOdrOffset.Start():emOdrOffset.End()]),
		BatPercent:  0.0,
	}
	if l >= emBatteryOffset.End() {
		eb.MicrophoneEvent.BatPercent = float32FromBytes(b[emBatteryOffset.Start():emBatteryOffset.End()])
	}
	return &eb
}

func newHallEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.HallEvent = &VehHallEvent{
		Battery:     float32FromBytes(b[ehBatteryOffset.Start():ehBatteryOffset.End()]),
		Temperature: float32FromBytes(b[ehTemperatureOffset.Start():ehTemperatureOffset.End()]),
		Odr:         float32FromBytes(b[ehOdrOffset.Start():ehOdrOffset.End()]),
		Hysteresis:  float32FromBytes(b[ehHysteresisOffset.Start():ehHysteresisOffset.End()]),
		Range:       float32FromBytes(b[ehRangeOffset.Start():ehRangeOffset.End()]),
		FirFilter:   binary.LittleEndian.Uint16(b[ehFirFilterOffset.Start():ehFirFilterOffset.End()]),
		BatPercent:  0.0,
	}
	if l >= ehBatteryOffset.End() {
		eb.MicrophoneEvent.BatPercent = float32FromBytes(b[ehBatteryOffset.Start():ehBatteryOffset.End()])
	}
	return &eb
}

func newConnectedEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.ConnectedEvent = &VehConnectedEvent{
		macString(b, l),
	}
	return &eb
}

func newDisconnectedEvent(b []byte) *VehEvent {
	eb, l := newVehEvent(b)
	eb.DisconnectedEvent = &VehDisconnectedEvent{
		macString(b, l),
	}
	return &eb
}

func macString(b []byte, l int) string {
	return fmt.Sprintf("%X:%X:%X:%X:%X:%X", b[l-1], b[l-2], b[l-3], b[l-4], b[l-5], b[l-6])
}

func float32FromBytes(b []byte) float32 {
	intVal := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(intVal)
}
