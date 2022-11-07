package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"time"
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

func (c *client) sendHeartbeat() {
	c.writer.WriteByte(heartbeatMsg)
	c.writer.Flush()
}

// Starts sending a heartbeat to the client at the defined interval.
// interval is expected to be in deciseconds (1/10th of a second).
func (c *client) startHeartbeat(interval uint32) {
	c.heartbeatTicker = time.NewTicker(time.Duration(interval*100) * time.Millisecond)
	c.heartbeatDoneChan = make(chan bool)

	go func() {
		for {
			select {
			case <-c.heartbeatDoneChan:
				return
			case t := <-c.heartbeatTicker.C:
				fmt.Println("Tick at", t)
				c.sendHeartbeat()
			}
		}
	}()
}
