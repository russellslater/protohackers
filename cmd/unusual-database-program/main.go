package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/russellslater/protohackers/internal/db"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	var host string
	flag.StringVar(&host, "host", "0.0.0.0", "Host address for server to bind to")
	flag.Parse()

	s := NewUnusualDatabaseServer(5000, host)
	defer s.Close()

	err := s.Start()
	if err != nil {
		log.Print(err)
		exitCode = 1
		return
	}
}

type UnusualDatabaseServer struct {
	port int
	host string
	conn net.PacketConn
	db   *db.UnusualDatabase
}

func NewUnusualDatabaseServer(port int, host string) *UnusualDatabaseServer {
	return &UnusualDatabaseServer{
		port: port,
		host: host,
		db:   db.NewUnusualDatabase(),
	}
}

func (s *UnusualDatabaseServer) Start() error {
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("can't listen on %d/udp: %s", s.port, err)
	}

	log.Printf("listening on %v", conn.LocalAddr())

	s.conn = conn
	s.handleUDP()

	return nil
}

func (s *UnusualDatabaseServer) Close() {
	s.conn.Close()
}

func (s *UnusualDatabaseServer) handleUDP() {
	buf := make([]byte, 1000)
	for {
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("error reading over UDP from %v: %s", addr, err)
			continue
		}

		log.Printf("received %d bytes over UDP from %v: %s", n, addr, buf[:n])

		response, send := s.handleCommand(string(bytes.Trim(buf[:n], "\x00")))
		if send {
			log.Printf("sending %d bytes over UDP to %v: %s", len(response), addr, response)
			s.conn.WriteTo([]byte(response), addr)
		}
	}
}

func (s *UnusualDatabaseServer) handleCommand(cmd string) (string, bool) {
	eqIndex := strings.Index(cmd, "=")
	if eqIndex != -1 {
		key := cmd[:eqIndex]
		value := cmd[eqIndex+1:]

		log.Printf("setting key %q to %q", key, value)

		s.db.Set(key, value)

		return "", false
	} else {
		value, _ := s.db.Get(cmd)
		return fmt.Sprintf("%s=%s", cmd, value), true
	}
}
