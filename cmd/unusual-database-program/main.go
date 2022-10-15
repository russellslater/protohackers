package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

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
	conn *net.UDPConn
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

func (s *UnusualDatabaseServer) Close() {
	s.conn.Close()
}

func (s *UnusualDatabaseServer) handleUDP() {
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

		result := NewDbCommand(s.db, string(bytes.Trim(buf[:n], "\x00"))).execute()

		if result != "" {
			log.Printf("sending %d bytes over UDP to %v: %s", len(result), addr, result)
			s.conn.WriteToUDP([]byte(result), addr)
		}
	}
}
