package main

import (
	"net"
	"strings"
	"sync"
)

type Session struct {
	ID          int
	Addr        *net.UDPAddr
	IsOpen      bool
	ReceivedPos int
	SentPos     int
	data        strings.Builder
	sync.Mutex
}

func NewSession(sid int, addr *net.UDPAddr) *Session {
	return &Session{
		ID:     sid,
		Addr:   addr,
		IsOpen: true,
	}
}

func (s *Session) AppendData(data string) int {
	s.Lock()
	defer s.Unlock()

	len, err := s.data.WriteString(data)
	if err == nil {
		s.ReceivedPos += len
		return s.ReceivedPos
	}

	return 0
}

func (s *Session) CompletedLines(pos int) ([]string, int) {
	data := s.data.String()

	if len(data) < pos {
		return nil, 0
	}

	idx := strings.LastIndex(data, "\n")

	if idx != -1 {
		return strings.Split(data[pos:idx], "\n"), idx - pos + 1
	} else {
		return nil, 0
	}
}
