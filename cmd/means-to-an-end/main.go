package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/russellslater/protohackers"
)

func main() {
	log.Fatal(protohackers.ListenAndAccept(5000, handle))
}

func handle(c net.Conn) error {
	defer c.Close()

	prices := make(map[int32]int32)

	// each message from a client is 9 bytes long
	buf := make([]byte, 9)

	for {
		n, err := io.ReadFull(c, buf)
		if err != nil || n != len(buf) {
			return err
		}

		t, arg1, arg2 := parseCommand(buf)
		res := executeCommand(t, arg1, arg2, prices)

		if res != nil {
			if _, err := c.Write(res); err != nil {
				return err
			}
		}
	}
}

func parseCommand(buf []byte) (rune, int32, int32) {
	t := rune(buf[0])
	arg1 := int32(binary.BigEndian.Uint32(buf[1:5]))
	arg2 := int32(binary.BigEndian.Uint32(buf[5:]))
	return t, arg1, arg2
}

func executeCommand(t rune, arg1 int32, arg2 int32, prices map[int32]int32) []byte {
	switch t {
	case 'I':
		insertPrice(arg1, arg2, prices)
	case 'Q':
		mean := queryPrice(arg1, arg2, prices)
		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs, uint32(mean))
		return bs
	}

	return nil
}

func insertPrice(timestamp int32, price int32, prices map[int32]int32) {
	prices[timestamp] = price
}

func queryPrice(mintime int32, maxtime int32, prices map[int32]int32) int32 {
	var total int64 // int64 to avoid overflow
	var count int64
	for time, p := range prices {
		if time >= mintime && time <= maxtime {
			total += int64(p)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	// integer division; "acceptable to round either up or down, at the server's discretion"
	return int32(total / count)
}
