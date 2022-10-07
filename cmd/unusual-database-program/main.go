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

	var address string
	flag.StringVar(&address, "address", "0.0.0.0", "IP address for server to bind to")
	flag.Parse()

	s := NewUnusualDatabaseServer(5000, net.ParseIP(address))
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
	addr := &net.UDPAddr{
		IP:   s.ip,
		Port: s.port,
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("can't listen on %d/udp: %s", addr.Port, err)
	}

	log.Printf("listening on %v", addr)

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

		response, send := s.handleCommand(string(bytes.Trim(buf[:n], "\x00")))
		if send {
			log.Printf("sending %d bytes over UDP to %v: %s", len(response), addr, response)
			s.conn.WriteToUDP([]byte(response), addr)
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
