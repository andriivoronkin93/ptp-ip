package ip

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/malc0mn/ptp-ip/ip/internal"
	"github.com/malc0mn/ptp-ip/ptp"
)

type DataPhase uint32
type PacketType uint32
type FailReason uint32
type ProtocolVersion uint32

const (
	HeaderSize int = 8

	// DP_NoDataOrDataIn is data being transferred from the Responder to the Initiator.
	DP_NoDataOrDataIn DataPhase = 0x00000001
	// DP_DataOut is data being transferred from the Initiator to the Responder.
	DP_DataOut DataPhase = 0x00000002
	// DP_Unknown indicates the Initiator does not know the data phase. The use of the "Unknown Data Phase" value in the
	// Operation Request Packet may significantly complicate the implementation of the protocol. When a packet with such
	// value is issued by the Initiator neither side of the communication "knows" what type of packet to expect next,
	// and, more importantly, in what direction. The knowledge may be available at higher levels of the communication
	// stack (PTP or application) or may not be available at all until the next packet arrives for transmission at a
	// later time. The latter is an important case of a "bridge" or "repeater" type of application that translates PTP
	// communication between USB and PTP-IP channels. In case of implementations for devices with limited processing
	// capabilities it may not always be possible to fully support this feature. In those cases, if the Responder cannot
	// handle the "Unknown Data Phase" value properly, it is recommended that it closes the connection upon receipt of
	// such packets.
	DP_Unknown DataPhase = 0x00000003

	FR_FailRejectedInitiator FailReason = 0x00000001
	FR_FailBusy              FailReason = 0x00000002
	FR_FailUnspecified       FailReason = 0x00000003

	// PKT_Invalid is not specified by the PTP/IP protocol. We use this to identify packets that deviate from the
	// standard. These will be treated differently when they are sent or received.
	PKT_Invalid            PacketType = 0x00000000
	PKT_InitCommandRequest PacketType = 0x00000001
	PKT_InitCommandAck     PacketType = 0x00000002
	PKT_InitEventRequest   PacketType = 0x00000003
	PKT_InitEventAck       PacketType = 0x00000004
	PKT_InitFail           PacketType = 0x00000005
	PKT_OperationRequest   PacketType = 0x00000006
	PKT_OperationResponse  PacketType = 0x00000007
	PKT_Event              PacketType = 0x00000008
	PKT_StartData          PacketType = 0x00000009
	PKT_Data               PacketType = 0x0000000A
	PKT_Cancel             PacketType = 0x0000000B
	PKT_EndData            PacketType = 0x0000000C
	PKT_ProbeRequest       PacketType = 0x0000000D
	PKT_ProbeResponse      PacketType = 0x0000000E

	PV_VersionOnePointZero ProtocolVersion = 0x00010000
)

var (
	UnknownPacketType = errors.New("unknown packet type %#x")
)

type Packet interface {
	PacketType() PacketType
}

type PacketOut interface {
	Packet
	Payload() []byte
}

type PacketIn interface {
	Packet
	TotalFixedFieldSize() int
}

type Header struct {
	Length     uint32
	PacketType PacketType
}

type InitCommandRequestPacket interface {
	PacketOut
	GetGUID() uuid.UUID
	GetFriendlyName() string
	GetProtocolVersion() ProtocolVersion
	SetProtocolVersion(pv ProtocolVersion)
}

// GenericInitCommandRequestPacket is used immediately after the Command/Data TCP ip channel is established. It is sent
// by the Initiator to the Responder on the Data/Command TCP connection and is used to communicate the identity of the
// Initiator to the Responder. The Responder can implement a filtering mechanism denying certain identities.
type GenericInitCommandRequestPacket struct {
	GUID uuid.UUID
	// A null terminated string.
	FriendlyName string
	// The 16 most significant bits are the major number, the 16 least significant bits are the minor number.
	ProtocolVersion ProtocolVersion
}

func (icrp *GenericInitCommandRequestPacket) PacketType() PacketType {
	return PKT_InitCommandRequest
}

func (icrp *GenericInitCommandRequestPacket) Payload() []byte {
	return internal.MarshalLittleEndian(icrp)
}

func (icrp *GenericInitCommandRequestPacket) GetGUID() uuid.UUID {
	return icrp.GUID
}

func (icrp *GenericInitCommandRequestPacket) GetFriendlyName() string {
	return icrp.FriendlyName
}

func (icrp *GenericInitCommandRequestPacket) GetProtocolVersion() ProtocolVersion {
	return icrp.ProtocolVersion
}

func (icrp *GenericInitCommandRequestPacket) SetProtocolVersion(pv ProtocolVersion) {
	icrp.ProtocolVersion = pv
}

func NewInitCommandRequestPacket(guid uuid.UUID, friendlyName string) InitCommandRequestPacket {
	return &GenericInitCommandRequestPacket{
		GUID:            guid,
		FriendlyName:    friendlyName,
		ProtocolVersion: PV_VersionOnePointZero,
	}
}

func NewInitCommandRequestPacketForClient(c *Client) InitCommandRequestPacket {
	return NewInitCommandRequestPacket(c.InitiatorGUID(), c.InitiatorFriendlyName())
}

func NewInitCommandRequestPacketWithVersion(guid uuid.UUID, friendlyName string, protocolVersion ProtocolVersion) InitCommandRequestPacket {
	icrp := NewInitCommandRequestPacket(guid, friendlyName)
	icrp.SetProtocolVersion(protocolVersion)

	return icrp
}

// InitCommandAckPacket is sent by the Responder in response to an InitCommandRequestPacket, to communicate the assigned
// ConnectionNumber for the PTP-IP session. It is transmitted on Data/Command TCP connection.
type InitCommandAckPacket struct {
	// A unique number generated by the Responder used to associate the TCP ip channels belonging to same PTP-IP
	// session. Reuse this number in the requests that will follow the InitCommandACKPacket.
	ConnectionNumber      uint32
	ResponderGUID         uuid.UUID
	ResponderFriendlyName string // null terminated string
	// The 16 most significant bits are the major number, the 16 least significant bits are the minor number.
	ResponderProtocolVersion uint32
}

func (icap *InitCommandAckPacket) PacketType() PacketType {
	return PKT_InitCommandAck
}

func (icap *InitCommandAckPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(icap)
}

type InitEventRequestPacket interface {
	PacketOut
	GetConnectionNumber() uint32
}

// GenericInitEventRequestPacket is used by the Initiator after the Command/Data TCP Connection is established, in order
// to establish the Event TCP Connection. When the Initiator receives a valid InitCommandAckPacket it establishes the
// Event TCP connection and transmits this packet on the Event TCP connection. The connection number received via the
// InitCommandAckPacket is reused in this packet.
type GenericInitEventRequestPacket struct {
	ConnectionNumber uint32
}

func (ierp *GenericInitEventRequestPacket) PacketType() PacketType {
	return PKT_InitEventRequest
}

func (ierp *GenericInitEventRequestPacket) Payload() []byte {
	return internal.MarshalLittleEndian(ierp)
}

func (ierp *GenericInitEventRequestPacket) GetConnectionNumber() uint32 {
	return ierp.ConnectionNumber
}

func NewInitEventRequestPacket(connNum uint32) InitEventRequestPacket {
	return &GenericInitEventRequestPacket{
		ConnectionNumber: connNum,
	}
}

// InitEventAckPacket is used by the Responder to inform the Initiator that the PTP-IP connection establishment has
// completed successfully. It is transmitted on the Event TCP connection.
type InitEventAckPacket struct {
}

func (ieap *InitEventAckPacket) PacketType() PacketType {
	return PKT_InitEventAck
}

func (ieap *InitEventAckPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(ieap)
}

// InitFailPacket is used by the Responder to inform the Initiator that the PTP-IP connection establishment failed. The
// reason of failure is reported in the Reason field. Upon receiving the packet, the Initiator MUST close the
// Command/Data TCP Connection with the Responder that rejects the event connection request. After issuing an
// InitFailPacket, the Responder SHALL close the PTP-IP connection (TCP connections initiated from the Initiator that
// has been rejected). The InitFailPacket can be transported on either of the TCP connections.
type InitFailPacket struct {
	Reason FailReason
}

func (ifp *InitFailPacket) PacketType() PacketType {
	return PKT_InitFail
}

func (ifp *InitFailPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(ifp)
}

func (ifp *InitFailPacket) ReasonAsError() error {
	var msg string
	switch ifp.Reason {
	case FR_FailBusy:
		msg = "busy: too many active connections"
	case FR_FailRejectedInitiator:
		msg = "rejected: device not allowed"
	case FR_FailUnspecified:
		msg = "reason unspecified"
	// TODO: should we not split off the vendor related errors somehow, to prevent this from becoming a very long list?
	case FR_Fuji_DeviceBusy:
		msg = "fuji: invalid friendly name or camera state: allow to 'change' client or 'reset' connection"
	case FR_Fuji_InvalidParameter:
		msg = "fuji: unknown protocol version"
	default:
		msg = fmt.Sprintf("unknown failure reason returned %#x", ifp.Reason)
	}

	return errors.New(msg)
}

// OperationRequestPacket is used to transport operation requests. PTP-IP Operation Request Packets are issued by the
// Initiator and are transported to the Responder device via the PTP-IP Command/Data ip channel. The direction of this
// packet is from Initiator to Responder.
// If the DataPhaseInfo field is set to DP_DataOut, then this packet MUST be followed by a StartDataPacket.
// If the Initiator wants to transfer a null data object to the Responder, than it has two options:
//   1. Set the DataPhaseInfo field to DP_NoDataOrDataIn, in which case the responder will follow up with an
//      OperationResponsePacket, without waiting for a data.
//   2. Set the DataPhaseInfo field to DP_DataOut. In this case, the data out phase MUST consist of exactly one
//      StartDataPacket, having the TotalDataLength field set to 0x00000000, and one empty EndDataPacket. The
//      Initiator MUST NOT send any other data packets.
type OperationRequestPacket struct {
	DataPhaseInfo DataPhase
	ptp.OperationRequest
}

func (orp *OperationRequestPacket) PacketType() PacketType {
	return PKT_OperationRequest
}

func (orp *OperationRequestPacket) Payload() []byte {
	return internal.MarshalLittleEndian(orp)
}

// OperationResponsePacket is used to transport Operation Responses by the Responder and are transported to the
// Initiator via the Command/Data TCP connection. PTP-IP Operation Response Packets are only issued by the Responder to
// indicate that the requested operation transaction has been completed and to pass the operation result.
type OperationResponsePacket struct {
	ptp.OperationResponse
}

func (orp *OperationResponsePacket) PacketType() PacketType {
	return PKT_OperationResponse
}

func (orp *OperationResponsePacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(orp)
}

type EventPacket interface {
	PacketIn
	GetEventCode() ptp.EventCode
}

type EventParameters struct {
	Parameter1 []byte
}

// GenericEventPacket is used to send PTP Events on the Event TCP connection. The events are used to inform the
// Initiator about the Responder state change.
type GenericEventPacket struct {
	ptp.Event
}

func (ep *GenericEventPacket) PacketType() PacketType {
	return PKT_Event
}

func (ep *GenericEventPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(ep)
}

func (ep *GenericEventPacket) GetEventCode() ptp.EventCode {
	return ep.EventCode
}

func NewEventPacket() EventPacket {
	return &GenericEventPacket{}
}

// StartDataPacket is used to signal the beginning of a data transfer. It is a is bi-directional packet, so this packet
// is either from the Responder to the Initiator or from the Initiator to the Responder. It is transmitted on the
// Command/Data TCP connection.
type StartDataPacket struct {
	TransactionId ptp.TransactionID
	// A value of 0xFFFFFFFFFFFFFFFF indicates that the size of the data is not known at the beginning of the data phase.
	TotalDataLength uint64
}

func (sdp *StartDataPacket) PacketType() PacketType {
	return PKT_StartData
}

func (sdp *StartDataPacket) Payload() []byte {
	return internal.MarshalLittleEndian(sdp)
}

func (sdp *StartDataPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(sdp)
}

// DataPacket is used to transport data. DataPackets are only used during data phase of a transaction and can be issued
// either by the Initiator or Responder in the direction of the data flow:
//   1. for the data-in phase - from the Responder to the Initiator
//   2. for the data-out phase - from the Initiator to the Responder.
// Data Packets are transmitted on Command/Data TCP connection.
// Normally there is no need in doing fragmentation and assembly of large data packets. However, a basic fragmentation
// mechanism MAY be utilized to allow for a simple data transfer cancelling mechanism. No error checking is required.
type DataPacket struct {
	TransactionId ptp.TransactionID
	DataPayload   interface{}
}

func (dp *DataPacket) PacketType() PacketType {
	return PKT_Data
}

func (dp *DataPacket) Payload() []byte {
	return internal.MarshalLittleEndian(dp)
}

func (dp *DataPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(dp)
}

// EndDataPacket is used to indicate the end of the data phase. The EndDataPacket can also carry useful data. This
// packet is only used during data phase of a transaction and can be issued either by the Initiator or Responder in the
// direction of the data flow: for the data-in phase - from the Responder to the Initiator; for the data-out phase -
// from the Initiator to the Responder.
type EndDataPacket struct {
	TransactionId ptp.TransactionID
	DataPayload   []byte
}

func (edp *EndDataPacket) PacketType() PacketType {
	return PKT_EndData
}

func (edp *EndDataPacket) Payload() []byte {
	return internal.MarshalLittleEndian(edp)
}

func (edp *EndDataPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(edp)
}

// CancelPacket is used to cancel a transaction.
type CancelPacket struct {
	TransactionId ptp.TransactionID
}

func (cp *CancelPacket) PacketType() PacketType {
	return PKT_Cancel
}

func (cp *CancelPacket) Payload() []byte {
	return internal.MarshalLittleEndian(cp)
}

func (cp *CancelPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(cp)
}

// ProbeRequestPacket can be used by both Initiator and Responder to check if a peer device is still active. Upon
// receiving such a packet, the device MUST respond immediately with a ProbeResponsePacket. If no response is received
// within a reasonable period of time, the device initiating this check will close the active PTP-IP session(s) with the
// remote device.
// This packet should be used with utmost care in order to avoid overloading of the LAN.
//   1. Initiator to Responder: it is recommended that this packet is used only during a PTP transaction (e.g. when a
//      format command is issued; if the storage media is large, the response time can be quite large), in order to
//      check out if the Responder is still active or not.
//   2. Responder to Initiator: it is recommended to use this packet only when the Responder receives a request for a
//      new PTP-IP session while one ore more other sessions are active. In this case, the Responder can check if the
//      existing PTP-IP connections are still active.
//   3. It is recommended that a timeout of 10 seconds be set between sending the Probe Request Packet and receiving the
//      Probe Response Packet.
type ProbeRequestPacket struct {
}

func (prqp *ProbeRequestPacket) PacketType() PacketType {
	return PKT_ProbeRequest
}

func (prqp *ProbeRequestPacket) Payload() []byte {
	return internal.MarshalLittleEndian(prqp)
}

func (prqp *ProbeRequestPacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(prqp)
}

// ProbeResponsePacket can be used in PTP-IP by both Initiator and Responder, as a response to a ProbeRequestPacket.
// Upon receiving a ProbeRequestPacket, a Probe Response Packet MUST be issued immediately. The Probe Response Packet is
// sent on the Event TCP connection.
type ProbeResponsePacket struct {
}

func (prsp *ProbeResponsePacket) PacketType() PacketType {
	return PKT_ProbeResponse
}

func (prsp *ProbeResponsePacket) Payload() []byte {
	return internal.MarshalLittleEndian(prsp)
}

func (prsp *ProbeResponsePacket) TotalFixedFieldSize() int {
	return internal.TotalSizeOfFixedFields(prsp)
}

// NewPacketOutFromPacketType creates an new packet struct based on the given packet type. All fields will be left
// uninitialised.
func NewPacketOutFromPacketType(pt PacketType) (PacketOut, error) {
	var p PacketOut

	switch pt {
	case PKT_InitCommandRequest:
		p = new(GenericInitCommandRequestPacket)
	case PKT_InitEventRequest:
		p = new(GenericInitEventRequestPacket)
	case PKT_OperationRequest:
		p = new(OperationRequestPacket)
	case PKT_StartData:
		p = new(StartDataPacket)
	case PKT_Data:
		p = new(DataPacket)
	case PKT_Cancel:
		p = new(CancelPacket)
	case PKT_EndData:
		p = new(EndDataPacket)
	case PKT_ProbeRequest:
		p = new(ProbeRequestPacket)
	case PKT_ProbeResponse:
		p = new(ProbeResponsePacket)
	}

	if p != nil {
		return p, nil
	}

	return nil, fmt.Errorf(UnknownPacketType.Error(), pt)
}

// NewPacketInFromPacketType creates an new packet struct based on the given packet type. All fields will be left
// uninitialised.
func NewPacketInFromPacketType(pt PacketType) (PacketIn, error) {
	var p PacketIn

	switch pt {
	case PKT_InitCommandAck:
		p = new(InitCommandAckPacket)
	case PKT_InitEventAck:
		p = new(InitEventAckPacket)
	case PKT_InitFail:
		p = new(InitFailPacket)
	case PKT_OperationResponse:
		p = new(OperationResponsePacket)
	case PKT_Event:
		p = new(GenericEventPacket)
	case PKT_StartData:
		p = new(StartDataPacket)
	case PKT_Data:
		p = new(DataPacket)
	case PKT_Cancel:
		p = new(CancelPacket)
	case PKT_EndData:
		p = new(EndDataPacket)
	case PKT_ProbeRequest:
		p = new(ProbeRequestPacket)
	case PKT_ProbeResponse:
		p = new(ProbeResponsePacket)
	}

	if p != nil {
		return p, nil
	}

	return nil, fmt.Errorf(UnknownPacketType.Error(), pt)
}
