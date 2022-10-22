package chatproxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Rewriter interface {
	Rewrite(string) string
	RewriteBytes([]byte) []byte
}

type ChatProxy struct {
	listenPort int
	remoteAddr string
	listener   net.Listener
	Rewriters  []Rewriter
}

func NewChatProxy(port int, remoteAddr string) *ChatProxy {
	return &ChatProxy{
		listenPort: port,
		remoteAddr: remoteAddr,
	}
}

func (s *ChatProxy) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.listenPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	log.Printf("listening on port %d\n", s.listenPort)

	for {
		client, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		go func() {
			if err := s.handle(client); err != nil {
				log.Printf("%s\n", err.Error())
			}
		}()
	}
}

func (s *ChatProxy) Close() {
	if err := s.listener.Close(); err != nil {
		log.Printf("error closing connection: %s\n", err.Error())
	}
}

func (s *ChatProxy) handle(client net.Conn) error {
	defer client.Close()

	upstream, err := net.Dial("tcp", s.remoteAddr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer upstream.Close()

	go func() {
		if err := s.proxy(upstream, client); err != nil {
			log.Printf("downstream failed: %s\n", err.Error())
		}
	}()

	if err := s.proxy(client, upstream); err != nil {
		return fmt.Errorf("upstream failed: %w", err)
	}

	return nil
}

func (s *ChatProxy) proxy(from net.Conn, to net.Conn) error {
	reader := bufio.NewReader(from)
	for {
		scanned, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		scanned = scanned[:len(scanned)-1] // trim newline

		log.Printf("received: %s", string(scanned))

		for _, rw := range s.Rewriters {
			scanned = rw.RewriteBytes(scanned)
		}

		scanned = append(scanned, []byte("\n")...)

		if _, err := to.Write(scanned); err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
	}
}
