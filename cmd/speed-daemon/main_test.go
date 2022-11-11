package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"testing"

	"github.com/matryer/is"
)

func startTestServer(conns ...net.Conn) {
	s := NewTicketServer(5000)

	for _, conn := range conns {
		client := s.connect(conn)
		go func() {
			s.serve(client)
		}()
	}
}

func sendIAmCamera(conn net.Conn, road uint16, mile uint16, limit uint16) {
	w := bufio.NewWriter(conn)
	w.WriteByte(0x80)

	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, road)
	w.Write([]byte(b))

	binary.BigEndian.PutUint16(b, mile)
	w.Write([]byte(b))

	binary.BigEndian.PutUint16(b, limit)
	w.Write([]byte(b))

	w.Flush()
}

func sendPlate(conn net.Conn, plate string, timestamp uint32) {
	w := bufio.NewWriter(conn)
	w.WriteByte(0x20)

	w.WriteByte(byte(len(plate)))
	w.WriteString(plate)

	b := make([]byte, 4)

	binary.BigEndian.PutUint32(b, timestamp)
	w.Write([]byte(b))

	w.Flush()
}

func sendIAmDispatcher(conn net.Conn, roads []uint16) {
	w := bufio.NewWriter(conn)
	w.WriteByte(0x81)

	w.WriteByte(uint8(len(roads)))

	b := make([]byte, 2)
	for i := 0; i < len(roads); i++ {
		binary.BigEndian.PutUint16(b, roads[i])
		w.Write([]byte(b))
	}

	w.Flush()
}

func readErrorMsg(conn net.Conn) string {
	r := bufio.NewReader(conn)
	r.ReadByte() // 0x10

	len, _ := r.ReadByte()

	buf := make([]byte, uint8(len))
	io.ReadFull(r, buf)

	return string(buf)
}

func TestTicketServerScenario(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	camera1ClientConn, camera1ServerConn := net.Pipe()
	camera2ClientConn, camera2ServerConn := net.Pipe()
	dispatcherClientConn, dispatcherServerConn := net.Pipe()

	startTestServer(camera1ServerConn, camera2ServerConn, dispatcherServerConn)

	sendIAmCamera(camera1ClientConn, 123, 8, 60)
	sendPlate(camera1ClientConn, "UN1X", 0)

	sendIAmCamera(camera2ClientConn, 123, 9, 60)
	sendPlate(camera2ClientConn, "UN1X", 45)

	sendIAmDispatcher(dispatcherClientConn, []uint16{123})

	readerDispatcher := bufio.NewReader(dispatcherClientConn)

	msg, _ := readerDispatcher.ReadByte()

	is.Equal(msg, byte(0x21)) // ticket message expected

	plateLen, _ := readerDispatcher.ReadByte()

	is.Equal(plateLen, byte(4)) // plate length does not match

	buf := make([]byte, plateLen)
	io.ReadFull(readerDispatcher, buf)

	is.Equal(buf, []byte("UN1X")) // plate does not match

	var road uint16
	binary.Read(readerDispatcher, binary.BigEndian, &road)

	is.Equal(road, uint16(123)) // road does not match

	var mileStart uint16
	binary.Read(readerDispatcher, binary.BigEndian, &mileStart)

	is.Equal(mileStart, uint16(8)) // mile start does not match

	var timestampStart uint32
	binary.Read(readerDispatcher, binary.BigEndian, &timestampStart)

	is.Equal(timestampStart, uint32(0)) // timestamp start does not match

	var mileEnd uint16
	binary.Read(readerDispatcher, binary.BigEndian, &mileEnd)

	is.Equal(mileEnd, uint16(9)) // mile end does not match

	var timestampEnd uint32
	binary.Read(readerDispatcher, binary.BigEndian, &timestampEnd)

	is.Equal(timestampEnd, uint32(45)) // timestamp end does not match

	var speed uint16
	binary.Read(readerDispatcher, binary.BigEndian, &speed)

	is.Equal(speed, uint16(8000)) // speed does not match

	camera1ClientConn.Close()
	camera2ClientConn.Close()
	dispatcherClientConn.Close()
}

func TestCameraAlreadyIdentified(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	// identify once
	sendIAmCamera(clientConn, 123, 8, 60)

	// identify twice
	sendIAmCamera(clientConn, 123, 8, 60)

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Already identified") // error message mismatch

	clientConn.Close()
}

func TestDispatcherAlreadyIdentified(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	// identify once
	sendIAmDispatcher(clientConn, []uint16{123})

	// identify twice
	sendIAmDispatcher(clientConn, []uint16{123})

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Already identified") // error message mismatch

	clientConn.Close()
}

func TestCameraDispatcherAlreadyIdentified(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	// identify as a camera
	sendIAmCamera(clientConn, 123, 8, 60)

	// identify a second time but as a dispatcher
	sendIAmDispatcher(clientConn, []uint16{123})

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Already identified") // error message mismatch

	clientConn.Close()
}

func TestDispatcherCameraAlreadyIdentified(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	// identify as a dispatcher
	sendIAmDispatcher(clientConn, []uint16{123})

	// identify a second time but as a camera
	sendIAmCamera(clientConn, 123, 8, 60)

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Already identified") // error message mismatch

	clientConn.Close()
}

func TestPlateMessageBeforeIdentifyingAsCamera(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	sendPlate(clientConn, "HELLO", 0)

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Client must identify as camera to observe plate") // error message mismatch

	clientConn.Close()
}

func TestPlateMessageFromDispatcher(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	sendIAmDispatcher(clientConn, []uint16{123})
	sendPlate(clientConn, "HELLO", 0)

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Client must identify as camera to observe plate") // error message mismatch

	clientConn.Close()
}

func TestUnknownMessage(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	clientConn, serverConn := net.Pipe()

	startTestServer(serverConn)

	unknownMessageByte := 0x99

	w := bufio.NewWriter(clientConn)
	w.WriteByte(uint8(unknownMessageByte))
	w.Flush()

	errorMsg := readErrorMsg(clientConn)

	is.Equal(errorMsg, "Unknown message") // error message mismatch

	clientConn.Close()
}
