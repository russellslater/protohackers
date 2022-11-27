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
