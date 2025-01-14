package ip

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/malc0mn/ptp-ip/ptp"
)

func TestNewDefaultInitiator(t *testing.T) {
	got, err := NewDefaultInitiator()
	if err != nil {
		t.Errorf("NewDefaultInitiator() err = %s; want <nil>", err)
	}
	if got.GUID == uuid.Nil {
		t.Errorf("NewDefaultInitiator() GUID = %s; want valid non-empty UUID", got.GUID)
	}
	if got.FriendlyName != InitiatorFriendlyName {
		t.Errorf("NewDefaultInitiator() Friendlyname = %s; want %s", got.FriendlyName, InitiatorFriendlyName)
	}
}

func TestNewInitiatorWithFriendlyName(t *testing.T) {
	got, err := NewInitiator("Friendly test", "")
	if err != nil {
		t.Errorf("NewInitiator() err = %s; want <nil>", err)
	}
	if got.GUID == uuid.Nil {
		t.Errorf("NewInitiator() GUID = %s; want valid non-empty UUID", got.GUID)
	}
	wantName := "Friendly test"
	if got.FriendlyName != wantName {
		t.Errorf("NewInitiator() Friendlyname = %s; want %s", got.FriendlyName, wantName)
	}
}

func TestNewResponder(t *testing.T) {
	got := NewResponder(DefaultVendor, DefaultIpAddress, 25740, 25741, 25742)
	wantV := ptp.VendorExtension(0)
	if got.Vendor != wantV {
		t.Errorf("NewResponder() Vendor = %#x; want %#x", got.Vendor, wantV)
	}
	if got.IpAddress != DefaultIpAddress {
		t.Errorf("NewResponder() IpAddress = %s; want %s", got.IpAddress, DefaultIpAddress)
	}
	wantP := uint16(25740)
	if got.CommandDataPort != wantP {
		t.Errorf("NewResponder() CommandDataPort = %d; want %d", got.CommandDataPort, wantP)
	}
	wantP = uint16(25741)
	if got.EventPort != wantP {
		t.Errorf("NewResponder() EventPort = %d; want %d", got.EventPort, wantP)
	}
	wantP = uint16(25742)
	if got.StreamerPort != wantP {
		t.Errorf("NewResponder() StreamerPort = %d; want %d", got.StreamerPort, wantP)
	}
	if got.GUID != uuid.Nil {
		t.Errorf("NewResponder() FriendlyName = %s; want <nil>", got.GUID)
	}
	if got.FriendlyName != "" {
		t.Errorf("NewResponder() FriendlyName = %s; want <nil>", got.FriendlyName)
	}
}

func TestNewClient(t *testing.T) {
	guid := "cf2407bc-4b4c-4525-9622-afb30db356df"
	got, err := NewClient(DefaultVendor, DefaultIpAddress, 26831, "", guid, logLevel)
	if err != nil {
		t.Errorf("NewClient() err = %s; want <nil>", err)
	}
	if got.CommandDataConn != nil {
		t.Errorf("NewClient() commandDataConn = %v; want <nil>", got.CommandDataConn)
	}
	if got.eventConn != nil {
		t.Errorf("NewClient() eventConn = %v; want <nil>", got.eventConn)
	}
	if got.initiator == nil {
		t.Errorf("NewClient() initiator = %v; want Initiator", got.initiator)
	}
	if got.responder == nil {
		t.Errorf("NewClient() responder = %v; want Responder", got.responder)
	}

	if got.ConnectionNumber() != 0 {
		t.Errorf("NewClient() ConnectionNumber() = %d; want 0", got.ConnectionNumber())
	}
	if got.TransactionId() != 0 {
		t.Errorf("NewClient() TransactionId() = %d; want 0", got.TransactionId())
	}
	want := "tcp"
	if got.Network() != want {
		t.Errorf("NewClient() Network() = %s; want %s", got.Network(), want)
	}
	want = "192.168.0.1:26831"
	if got.CommandDataAddress() != want {
		t.Errorf("NewClient() CommandDataAddress() = %s; want %s", got.CommandDataAddress(), want)
	}
	if got.EventAddress() != want {
		t.Errorf("NewClient() EventAddress() = %s; want %s", got.EventAddress(), want)
	}
	if got.StreamerAddress() != want {
		t.Errorf("NewClient() StreamerAddress() = %s; want %s", got.StreamerAddress(), want)
	}
	want = ""
	if got.ResponderFriendlyName() != want {
		t.Errorf("NewClient() ResponderFriendlyName() = %s; want %s", got.ResponderFriendlyName(), want)
	}
	want = "Golang PTP/IP client"
	if got.InitiatorFriendlyName() != want {
		t.Errorf("NewClient() InitiatorFriendlyName() = %s; want %s", got.InitiatorFriendlyName(), want)
	}
	wantv := ptp.VendorExtension(0)
	if got.ResponderVendor() != wantv {
		t.Errorf("NewClient() ResponderVendor() = %#x; want %#x", got.ResponderVendor(), wantv)
	}
	wantg := uuid.Nil
	if got.ResponderGUID() != wantg {
		t.Errorf("NewClient() ResponderGUID() = %s; want %s", got.ResponderGUID(), wantg)
	}
	want = "00000000-0000-0000-0000-000000000000"
	if got.ResponderGUIDAsString() != want {
		t.Errorf("NewClient() ResponderGUIDAsString() = %s; want %s", got.ResponderGUIDAsString(), want)
	}
	wantg, _ = uuid.Parse(guid)
	if got.InitiatorGUID() != wantg {
		t.Errorf("NewClient() InitiatorGUID() = %s; want %s", got.InitiatorGUID(), wantg)
	}
	if got.InitiatorGUIDAsString() != guid {
		t.Errorf("NewClient() InitiatorGUIDAsString() = %s; want %s", got.InitiatorGUIDAsString(), guid)
	}
}

func TestClient_SetCommandDataPort(t *testing.T) {
	got, err := NewClient(DefaultVendor, DefaultIpAddress, 55286, "", "5d5069bd-57a5-46e2-83cc-63c897ace234", logLevel)
	if err != nil {
		t.Fatal(err)
	}

	want := "192.168.0.1:55286"
	if got.CommandDataAddress() != want {
		t.Errorf("NewClient() CommandDataAddress() = %s; want %s", got.CommandDataAddress(), want)
	}

	got.SetCommandDataPort(55740)
	want = "192.168.0.1:55740"
	if got.CommandDataAddress() != want {
		t.Errorf("NewClient() CommandDataAddress() = %s; want %s", got.CommandDataAddress(), want)
	}
}

func TestClient_SetEventPort(t *testing.T) {
	got, err := NewClient(DefaultVendor, DefaultIpAddress, 55348, "", "5d5069bd-57a5-46e2-83cc-63c897ace234", logLevel)
	if err != nil {
		t.Fatal(err)
	}

	want := "192.168.0.1:55348"
	if got.EventAddress() != want {
		t.Errorf("NewClient() EventAddress() = %s; want %s", got.EventAddress(), want)
	}

	got.SetEventPort(55741)
	want = "192.168.0.1:55741"
	if got.EventAddress() != want {
		t.Errorf("NewClient() EventAddress() = %s; want %s", got.EventAddress(), want)
	}
}

func TestClient_SetStreamerPort(t *testing.T) {
	got, err := NewClient(DefaultVendor, DefaultIpAddress, 51986, "", "5d5069bd-57a5-46e2-83cc-63c897ace234", logLevel)
	if err != nil {
		t.Fatal(err)
	}

	want := "192.168.0.1:51986"
	if got.StreamerAddress() != want {
		t.Errorf("NewClient() StreamerAddress() = %s; want %s", got.StreamerAddress(), want)
	}

	got.SetStreamerPort(55742)
	want = "192.168.0.1:55742"
	if got.StreamerAddress() != want {
		t.Errorf("NewClient() StreamerAddress() = %s; want %s", got.StreamerAddress(), want)
	}
}

func TestClient_incrementTransactionId(t *testing.T) {
	c := Client{}

	got := c.TransactionId()
	want := ptp.TransactionID(0)
	if got != want {
		t.Errorf("TransactionId() = %#x; want %#x", got, want)
	}

	c.incrementTransactionId()
	got = c.TransactionId()
	want = ptp.TransactionID(1)
	if got != want {
		t.Errorf("TransactionId() = %#x; want %#x", got, want)
	}

	got = c.incrementTransactionId()
	want = ptp.TransactionID(2)
	if got != want {
		t.Errorf("TransactionId() = %#x; want %#x", got, want)
	}

	c.transactionId = 0xFFFFFFFE
	c.incrementTransactionId()
	got = c.TransactionId()
	want = ptp.TransactionID(1)
	if got != want {
		t.Errorf("TransactionId() = %#x; want %#x", got, want)
	}
}

func TestClient_sendPacket(t *testing.T) {
	c, err := NewClient(DefaultVendor, DefaultIpAddress, DefaultPort, "writèr", "e462b590-b516-474a-9db8-a465b370fabd", logLevel)
	if err != nil {
		t.Errorf("sendPacket() err = %s; want <nil>", err)
	}

	p := NewInitCommandRequestPacketForClient(c)

	want := "[00101010 00000000 00000000 00000000 00000001 00000000 00000000 00000000 11100100 01100010 10110101 10010000 10110101 00010110 01000111 01001010 10011101 10111000 10100100 01100101 10110011 01110000 11111010 10111101 01110111 00000000 01110010 00000000 01101001 00000000 01110100 00000000 11101000 00000000 01110010 00000000 00000000 00000000 00000000 00000000 00000001 00000000]"

	var buf bytes.Buffer
	c.sendPacket(&buf, p)
	got := fmt.Sprintf("%.8b", buf.Bytes())

	if got != want {
		t.Errorf("sendPacket() buffer = %s; want %s", got, want)
	}
}

func TestClient_readResponse(t *testing.T) {
	c, err := NewClient(DefaultVendor, DefaultIpAddress, DefaultPort, "writèr", "d6555687-a599-44b8-a4af-279d599a92f6", logLevel)
	if err != nil {
		t.Errorf("readResponse() err = %s; want <nil>", err)
	}

	guidR, _ := uuid.Parse("7c946ae4-6d6a-4589-90ed-d059f8cc426b")

	var b bytes.Buffer
	sendAnyPacket(
		&b,
		&InitCommandAckPacket{uint32(1), guidR, "remôte", uint32(0x00020005)},
		nil,
		"[ip_test]",
	)

	rp, xs, err := c.readResponse(&b, nil)
	if len(xs) > 0 {
		t.Errorf("readResponse() excess bytes = %d; want 0", len(xs))
	}
	if err != nil {
		t.Errorf("readResponse() error = %s; want <nil>", err)
	}

	want := "*ip.InitCommandAckPacket"
	if fmt.Sprintf("%T", rp) != want {
		t.Errorf("readResponse() PaketType = %T; want %s", rp, want)
	}

	gotType := rp.PacketType()
	wantType := PKT_InitCommandAck
	if gotType != wantType {
		t.Errorf("readResponse() PaketType = %X; want %X", gotType, wantType)
	}

	gotNum := rp.(*InitCommandAckPacket).ConnectionNumber
	wantNum := uint32(1)
	if gotNum != wantNum {
		t.Errorf("readResponse() ConnectionNumber = %d; want %d", gotNum, wantNum)
	}

	gotGuid := rp.(*InitCommandAckPacket).ResponderGUID
	wantGuid, _ := uuid.Parse("7c946ae4-6d6a-4589-90ed-d059f8cc426b")
	if gotGuid != wantGuid {
		t.Errorf("readResponse() ResponderGUID = %s; want %s", gotGuid, wantGuid)
	}

	gotName := rp.(*InitCommandAckPacket).ResponderFriendlyName
	wantName := "remôte"
	if gotName != wantName {
		t.Errorf("readResponse() ResponderFriendlyName = %s (%#x); want %s (%#x)", gotName, gotName, wantName, wantName)
	}

	gotVer := rp.(*InitCommandAckPacket).ResponderProtocolVersion
	wantVer := uint32(0x00020005)
	if gotVer != wantVer {
		t.Errorf("readResponse() ResponderProtocolVersion = %#x; want %#x", gotVer, wantVer)
	}
}

func TestClient_readRawResponse(t *testing.T) {
	c, err := NewClient(DefaultVendor, DefaultIpAddress, DefaultPort, "wrîter", "617b38ef-b6e6-4ef6-b2ad-ea51cecdbbd3", logLevel)
	if err != nil {
		t.Errorf("readRawResponse() err = %s; want <nil>", err)
	}

	guidR, _ := uuid.Parse("d2d4fce6-1181-42dd-a185-5cc40ca68321")

	var b bytes.Buffer
	sendAnyPacket(
		&b,
		&InitCommandAckPacket{uint32(1), guidR, "rèmote", uint32(0x00020005)},
		nil,
		"[ip_test]",
	)

	got, err := c.readRawResponse(&b)
	if err != nil {
		t.Errorf("readRawResponse() error = %s; want <nil>", err)
	}

	want := []byte{0x2e, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0xd2, 0xd4, 0xfc, 0xe6, 0x11, 0x81, 0x42, 0xdd, 0xa1, 0x85, 0x5c, 0xc4, 0xc, 0xa6, 0x83, 0x21, 0x72, 0x0, 0xe8, 0x0, 0x6d, 0x0, 0x6f, 0x0, 0x74, 0x0, 0x65, 0x0, 0x0, 0x0, 0x5, 0x0, 0x2, 0x0}
	if bytes.Compare(got, want) != 0 {
		t.Errorf("readRawResponse() raw = %v; want %v", got, want)
	}
}

func TestClient_initCommandDataConn(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, okPort, "testèr", "67bace55-e7a4-4fbc-8e31-5122ee73a17c", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = c.initCommandDataConn()
	if err != nil {
		t.Errorf("initCommandDataConn() error = %s; want <nil>", err)
	}

	got := c.TransactionId()
	want := ptp.TransactionID(0)
	if got != want {
		t.Errorf("TransactionId() got = %#x; want %#x", got, want)
	}
}
func TestClient_initCommandDataConnFail(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, failPort, "testér", "b3ca53e9-bb61-4c85-9fcd-3b446a9e81e6", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = c.initCommandDataConn()
	if err == nil {
		t.Errorf("initCommandDataConn() error = %s; want rejected: device not allowed", err)
	}

	got := c.TransactionId()
	want := ptp.TransactionID(0)
	if got != want {
		t.Errorf("TransactionId() got = %#x; want %#x", got, want)
	}
}

func TestClient_subscribe(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, failPort, "testér", "b3ca53e9-bb61-4c85-9fcd-3b446a9e81e6", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}

	tid := ptp.TransactionID(55)
	ch := make(chan []byte, 2)
	err = c.subscribe(tid, ch)
	if err != nil {
		t.Errorf("subscribe() error = %s; want: nil", err)
	}

	got, ok := c.cmdDataSubs[tid]
	if !ok {
		t.Errorf("subscribe() got = %#v; want true", got)
	}
	if got != ch {
		t.Errorf("subscribe() got = %#v; want %#v", got, ch)
	}
}

func TestClient_initEventConn(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, okPort, "testèr", "67bace55-e7a4-4fbc-8e31-5122ee73a17c", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = c.initEventConn()
	if err != nil {
		t.Errorf("initEventConn() error = %s; want <nil>", err)
	}

	got := c.TransactionId()
	want := ptp.TransactionID(1)
	if got != want {
		t.Errorf("TransactionId() got = %#x; want %#x", got, want)
	}
}

func TestClient_initEventConnFail(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, failPort, "testér", "733e8d71-0f05-4aba-9745-ea9294dd2278", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = c.initEventConn()
	if err == nil {
		t.Errorf("initEventConn() error = %s; want rejected: device not allowed", err)
	}

	got := c.TransactionId()
	want := ptp.TransactionID(0)
	if got != want {
		t.Errorf("TransactionId() got = %#x; want %#x", got, want)
	}
}

func TestClient_Dial(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, okPort, "testèr", "7e5ac7d3-46b7-4c50-b0d9-ba56c0e599f0", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = c.Dial()
	if err != nil {
		t.Errorf("Dial() err = %s; want <nil>", err)
	}

	c, err = NewClient(DefaultVendor, address, failPort, "testér", "f62b41f8-a094-4dab-b537-99afd04c6024", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = c.Dial()
	if err == nil {
		t.Errorf("Dial() err = %s; want rejected: device not allowed", err)
	}
}

func TestClient_GetDeviceInfo(t *testing.T) {
	c, err := NewClient(DefaultVendor, address, okPort, "tèster", "558acd44-f794-4b26-9129-d460b2a29e8d", logLevel)
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = c.Dial()
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.GetDeviceInfo()
	if err != nil {
		t.Errorf("GetDeviceInfo() err = %s; want <nil>", err)
	}
	if got == nil {
		t.Errorf("GetDeviceInfo() got = %v; want *ip.OperationResponsePacket", got)
	}
}
