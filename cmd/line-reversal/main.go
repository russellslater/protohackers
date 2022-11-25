package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/russellslater/protohackers/cmd/line-reversal/lrcpmsg"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	var host string
	flag.StringVar(&host, "host", "0.0.0.0", "Host address for server to bind to")
	flag.Parse()

	s := NewLineReversalServer(5000, host)
	defer s.Close()

	err := s.Start()
	if err != nil {
		log.Print(err)
		exitCode = 1
		return
	}
}

type LineReversalServer struct {
	port int
	host string
	conn *net.UDPConn
}

func NewLineReversalServer(port int, host string) *LineReversalServer {
	return &LineReversalServer{
		port: port,
		host: host,
	}
}

func (s *LineReversalServer) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s:%d: %s", s.host, s.port, err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("can't listen on %d/udp: %s", s.port, err)
	}

	log.Printf("listening on %v", conn.LocalAddr())

	s.conn = conn
	s.handleUDP()

	return nil
}

func (s *LineReversalServer) Close() {
	s.conn.Close()
}

func (s *LineReversalServer) handleUDP() {
	buf := make([]byte, 1000)
	for {
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("error reading over UDP from %v: %s", addr, err)
			continue
		}

		log.Printf("received %d bytes over UDP from %v: %s", n, addr, buf[:n])

		result, err := lrcpmsg.ParseMsg(bytes.Trim(buf[:n], "\x00"))

		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		var response string

		switch res := result.(type) {
		case lrcpmsg.ConnectMsg:
			response = fmt.Sprintf("Connect - Session ID: %d\n", res.SessionID)
		case lrcpmsg.DataMsg:
			response = fmt.Sprintf("Data - SessionID: %d, Pos: %d, %v\n", res.SessionID, res.Pos, res.Data)
		case lrcpmsg.AckMsg:
			response = fmt.Sprintf("Ack - Session ID: %d, Length: %d\n", res.SessionID, res.Length)
		case lrcpmsg.CloseMsg:
			response = fmt.Sprintf("Close - Session ID: %d\n", res.SessionID)
		}

		if response != "" {
			log.Printf("sending %d bytes over UDP to %v: %s", len(response), addr, response)
			s.conn.WriteToUDP([]byte(response), addr)
		}
	}
}
