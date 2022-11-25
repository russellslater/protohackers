package main

import "net"

type Session struct {
	ID          int
	Addr        *net.UDPAddr
	IsOpen      bool
	ReceivedPos int
	SentPos     int
}
