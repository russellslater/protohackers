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
	"sync"

	"github.com/russellslater/protohackers/cmd/line-reversal/lrcpmsg"
	"github.com/russellslater/protohackers/cmd/line-reversal/util"
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

	session = NewSession(sid, addr)

	s.sessions[sid] = session

	return session
}

func (s *LineReversalServer) findSession(sid int) *Session {
	session, ok := s.sessions[sid]

	if !ok {
		return nil
	}

	if !session.IsOpen {
		s.sendCloseMessage(session)
		return nil
	}

	return session
}

func (s *LineReversalServer) closeSession(sid int) *Session {
	s.Lock()
	defer s.Unlock()

	session, ok := s.sessions[sid]
	if ok {
		session.Close()
		s.sendCloseMessage(session)
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

		switch res := result.(type) {
		case lrcpmsg.ConnectMsg:
			session := s.openSession(res.SessionID, addr)
			if session.IsOpen {
				s.sendAckMessage(session, 0)
			} else {
				s.closeSession(session.ID)
			}
		case lrcpmsg.DataMsg:
			log.Printf("DataMsg - Session ID: %d, Pos: %d\n", res.SessionID, res.Pos)
			s.handleData(res)
		case lrcpmsg.AckMsg:
			s.handleAck(res)
		case lrcpmsg.CloseMsg:
			s.closeSession(res.SessionID)
		}
	}
}

func (s *LineReversalServer) handleData(msg lrcpmsg.DataMsg) {
	session := s.findSession(msg.SessionID)
	if session == nil {
		return
	}

	log.Printf("Session - ID: %d, ReceivedPos: %d\n", session.ID, session.ReceivedPos)
	log.Printf("Data - Pos: %d, %v\n", msg.Pos, msg.Data)

	if session.ReceivedPos == msg.Pos {
		data := util.SlashUnescape(string(msg.Data))
		log.Printf("Unescaped: %s, %d\n", data, len(data))

		session.AppendData(data)
		s.sendAckMessage(session, session.ReceivedPos)

		lines, len := session.CompletedLines(session.SentPos)

		if len > 0 {
			for i, l := range lines {
				lines[i] = string(util.Reverse([]byte(l)))
			}

			// TODO: handle retransmission (3s)
			// TODO: handle session expiry (60s)
			// TODO: LRCP messages must be smaller than 1000 bytes. You might have to break up data into multiple data messages in order to fit it below this limit.
			s.sendDataMessage(session, session.SentPos, lines)
			session.SentPos += len
		}
	} else if session.ReceivedPos < msg.Pos {
		s.sendAckMessage(session, session.ReceivedPos)
	}
}

func (s *LineReversalServer) handleAck(msg lrcpmsg.AckMsg) {
	session := s.findSession(msg.SessionID)
	if session == nil {
		return
	}

	log.Printf("Session - ID: %d, ReceivedPos: %d, LargestAckPos: %d\n", session.ID, session.SentPos, session.LargestAckPos)
	log.Printf("Ack - Length: %d\n", msg.Length)

	if msg.Length < session.LargestAckPos {
		return // do nothing
	}

	if msg.Length > session.SentPos {
		// misbehaving
		s.closeSession(session.ID)
	} else if msg.Length < session.SentPos {
		lines, len := session.CompletedLines(msg.Length)

		if len > 0 {
			for i, l := range lines {
				lines[i] = string(util.Reverse([]byte(l)))
			}

			// TODO: LRCP messages must be smaller than 1000 bytes.
			// You might have to break up data into multiple data messages in order to fit it below this limit.
			s.sendDataMessage(session, msg.Length, lines)
		}
	} else {
		session.LargestAckPos = msg.Length
	}
}

func (s *LineReversalServer) sendCloseMessage(session *Session) {
	s.sendMessage(session, lrcpmsg.CloseMsg{SessionID: session.ID})
}

func (s *LineReversalServer) sendAckMessage(session *Session, pos int) {
	ack := lrcpmsg.AckMsg{SessionID: session.ID, Length: pos}
	s.sendMessage(session, ack)
}

func (s *LineReversalServer) sendDataMessage(session *Session, pos int, lines []string) {
	data := util.SlashEscape(fmt.Sprintf("%s\n", strings.Join(lines, "\n")))
	msg := lrcpmsg.DataMsg{SessionID: session.ID, Pos: pos, Data: []byte(data)}
	s.sendMessage(session, msg)
}

func (s *LineReversalServer) sendMessage(session *Session, msg lrcpmsg.Msg) {
	log.Printf("sending %d bytes over UDP to %v: %s", len(fmt.Sprintf("%v", msg)), session.Addr, msg)
	s.conn.WriteToUDP([]byte(fmt.Sprintf("%v", msg)), session.Addr)
}
