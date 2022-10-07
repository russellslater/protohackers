package main

import (
	"bytes"
	"errors"
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

	// Required for fly.io
	// net.ParseIP("fly-global-services")

	s := NewUnusualDatabaseServer(5000, net.ParseIP("127.0.0.1"))
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
	ip   net.IP
	conn *net.UDPConn
	db   *db.UnusualDatabase
}

func NewUnusualDatabaseServer(port int, ip net.IP) *UnusualDatabaseServer {
	return &UnusualDatabaseServer{
		port: port,
		ip:   ip,
		db:   db.NewUnusualDatabase(),
	}
}

func (s *UnusualDatabaseServer) Start() error {
	addr := net.UDPAddr{
		IP:   s.ip,
		Port: s.port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("can't listen on %d/udp: %s", addr.Port, err)
	}

	log.Println("listening on port", addr.Port)

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
		n, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("error reading over UDP: %s", err)
			continue
		}

		log.Printf("received %d bytes over UDP: %s", n, buf[:n])

		s.handleCommand(string(bytes.Trim(buf[:n], "\x00")))
	}
}

func (s *UnusualDatabaseServer) handleCommand(cmd string) {
	eqIndex := strings.Index(cmd, "=")
	if eqIndex != -1 {
		key := cmd[:eqIndex]
		value := cmd[eqIndex+1:]

		log.Printf("setting key %q to %q", key, value)

		s.db.Set(key, value)
	}
}
