package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

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
	port     int
	host     string
	conn     *net.UDPConn
	sessions map[int]*Session
	sync.Mutex
}

func NewLineReversalServer(port int, host string) *LineReversalServer {
	return &LineReversalServer{
		port:     port,
		host:     host,
		sessions: make(map[int]*Session),
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

func (s *LineReversalServer) openSession(sid int, addr *net.UDPAddr) *Session {
	s.Lock()
	defer s.Unlock()

	session, ok := s.sessions[sid]
	if ok {
		return session
	}

	session = &Session{
		ID:     sid,
		Addr:   addr,
		IsOpen: true,
	}

	s.sessions[sid] = session

	return session
}

func (s *LineReversalServer) closeSession(sid int) *Session {
	s.Lock()
	defer s.Unlock()

	session, ok := s.sessions[sid]
	if ok {
		session.IsOpen = false
		return session
	}

	return nil
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
			session := s.openSession(res.SessionID, addr)
			if session.IsOpen {
				s.sendAckMessage(session)
			} else {
				s.sendCloseMessage(session)
			}
		case lrcpmsg.DataMsg:
			response = fmt.Sprintf("Data - SessionID: %d, Pos: %d, %v\n", res.SessionID, res.Pos, res.Data)
		case lrcpmsg.AckMsg:
			response = fmt.Sprintf("Ack - Session ID: %d, Length: %d\n", res.SessionID, res.Length)
		case lrcpmsg.CloseMsg:
			session := s.closeSession(res.SessionID)
			if session != nil {
				s.sendCloseMessage(session)
			}
		}

		log.Println(response)

		if response != "" {
			log.Printf("sending %d bytes over UDP to %v: %s", len(fmt.Sprintf("%v", response)), addr, response)
			s.conn.WriteToUDP([]byte(fmt.Sprintf("%v", response)), addr)
		}
	}
}

func (s *LineReversalServer) sendCloseMessage(session *Session) {
	s.sendMessage(session, lrcpmsg.CloseMsg{SessionID: session.ID})
}

func (s *LineReversalServer) sendAckMessage(session *Session) {
	ack := lrcpmsg.AckMsg{SessionID: session.ID, Length: session.ReceivedPos}
	s.sendMessage(session, ack)
}

func (s *LineReversalServer) sendMessage(session *Session, msg lrcpmsg.Msg) {
	log.Printf("sending %d bytes over UDP to %v: %s", len(fmt.Sprintf("%v", msg)), session.Addr, msg)
	s.conn.WriteToUDP([]byte(fmt.Sprintf("%v", msg)), session.Addr)
}
