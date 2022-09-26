package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	s := NewChatServer(5000)
	log.Fatal(s.Start())
}

type ChatServer struct {
	port     int
	clients  []*client
	listener net.Listener
}

type client struct {
	name string
	addr string
	conn net.Conn
}

func NewChatServer(port int) *ChatServer {
	return &ChatServer{
		port: port,
	}
}

func (s *ChatServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	fmt.Println("listening on port", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		client := s.connect(conn)

		go func() {
			if err := s.serve(client); err != nil {
				fmt.Println(err.Error())
			}
		}()
	}
}

func (s *ChatServer) connect(conn net.Conn) *client {
	client := &client{
		addr: conn.RemoteAddr().String(),
		conn: conn,
	}

	s.clients = append(s.clients, client)

	fmt.Printf("connection from %s [# connected clients: %d]\n", client.addr, len(s.clients))

	return client
}

func (s *ChatServer) remove(client *client) {
	for i, c := range s.clients {
		if client == c {
			s.clients[i] = s.clients[len(s.clients)-1]
			s.clients = s.clients[:len(s.clients)-1]
			break
		}
	}

	fmt.Printf("connection from %s closed\n", client.addr)

	client.conn.Close()
}

func (c *ChatServer) serve(client *client) error {
	defer c.remove(client)

	client.conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		line := string(scanner.Bytes())

		log.Println("received:", line)

		if client.name == "" {
			if c.validateClientName(line) {
				client.name = line
			} else {
				return fmt.Errorf("invalid client name: %s", line)
			}
		}
	}

	return scanner.Err()
}

func (c *ChatServer) validateClientName(name string) bool {
	// must contain at least one character
	if len(name) < 1 {
		return false
	}

	// must contain only alphanumeric characters
	for _, r := range strings.ToLower(name) {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}

	// must be a unique name
	for _, client := range c.clients {
		if name == client.name {
			return false
		}
	}

	return true
}

func (s *ChatServer) Close() {
	if err := s.listener.Close(); err != nil {
		fmt.Print("error closing connection: %w", err)
	}
}
