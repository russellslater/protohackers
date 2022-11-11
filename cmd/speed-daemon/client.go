package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/russellslater/protohackers/cmd/speed-daemon/ticketer"
)

type client struct {
	addr       string
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	camera     *camera
	dispatcher *dispatcher

	heartbeatTicker   *time.Ticker
	heartbeatDoneChan chan bool
}

type camera struct {
	road  uint16
	mile  uint16
	limit uint16
}

type dispatcher struct {
	roads []uint16
}

func (c *client) ID() string {
	return c.addr
}

func (c *client) Roads() []ticketer.RoadID {
	if c.isDispatcher() {
		roadIDs := make([]ticketer.RoadID, len(c.dispatcher.roads))
		for i := range c.dispatcher.roads {
			roadIDs[i] = ticketer.RoadID(c.dispatcher.roads[i])
		}
		return roadIDs
	}
	return []ticketer.RoadID{}
}

func (c *client) SendTicket(t *ticketer.Ticket) {
	if c.isDispatcher() {
		log.Printf("%v\n", t)

		c.writer.WriteByte(ticketMsg)
		c.writeString(t.Plate)
		c.writeUint16(uint16(t.Road))
		c.writeUint16(t.MileStart)
		c.writeUint32(t.TimestampStart)
		c.writeUint16(t.MileEnd)
		c.writeUint32(t.TimestampEnd)
		c.writeUint16(t.Speed)
		c.writer.Flush()
	}
}

func (c *client) isCamera() bool {
	return c.camera != nil
}

func (c *client) isDispatcher() bool {
	return c.dispatcher != nil
}

func (c *client) isIdentified() bool {
	return c.isCamera() || c.isDispatcher()
}

func (c *client) readMsg() (msgType, error) {
	m, err := c.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return msgType(m), nil
}

func (c *client) readUint16() uint16 {
	var num uint16
	// TODO: handle errors
	binary.Read(c.reader, binary.BigEndian, &num)
	return num
}

func (c *client) readUint16Array() []uint16 {
	count, err := c.reader.ReadByte()
	if err != nil {
		return []uint16{}
	}

	arr := make([]uint16, uint8(count))
	for i := 0; i < len(arr); i++ {
		arr[i] = c.readUint16()
	}

	return arr
}

func (c *client) readUint32() uint32 {
	var num uint32
	// TODO: handle errors
	binary.Read(c.reader, binary.BigEndian, &num)
	return num
}

func (c *client) readStr() string {
	len, err := c.reader.ReadByte()
	if err != nil {
		return ""
	}

	buf := make([]byte, uint8(len))
	_, err = io.ReadFull(c.reader, buf)
	if err != nil {
		return ""
	}

	return string(buf)
}

func (c *client) writeUint16(i uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	c.writer.Write(b)
}

func (c *client) writeUint32(i uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	c.writer.Write(b)
}

func (c *client) writeString(str string) {
	c.writer.WriteByte(byte(len(str)))
	c.writer.WriteString(str)
}

func (c *client) sendHeartbeat() {
	c.writer.WriteByte(heartbeatMsg)
	c.writer.Flush()
}

func (c *client) sendError(msg string) {
	if len(msg) <= math.MaxUint8 {
		c.writer.WriteByte(errorMsg)
		c.writeString(msg)
		c.writer.Flush()
	}
}

// Starts sending a heartbeat to the client at the defined interval.
// interval is expected to be in deciseconds (1/10th of a second).
func (c *client) startHeartbeat(interval uint32) {
	if interval == 0 {
		return
	}

	c.heartbeatTicker = time.NewTicker(time.Duration(interval*100) * time.Millisecond)
	c.heartbeatDoneChan = make(chan bool)

	go func() {
		for {
			select {
			case <-c.heartbeatDoneChan:
				return
			case <-c.heartbeatTicker.C:
				c.sendHeartbeat()
			}
		}
	}()
}

func (c *client) stopHeartbeat() {
	c.heartbeatDoneChan <- true
	c.heartbeatTicker.Stop()
}

func (c *client) isHeartbeatEnabled() bool {
	return c.heartbeatTicker != nil
}
