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
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	log.Println("listening on port", s.port)

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

func (s *ChatServer) Close() {
	if err := s.listener.Close(); err != nil {
		fmt.Print("error closing connection: %w", err)
	}
}

func (s *ChatServer) connect(conn net.Conn) *client {
	client := &client{
		addr: conn.RemoteAddr().String(),
		conn: conn,
	}

	s.clients = append(s.clients, client)

	log.Printf("connection from %s [# connected clients: %d]\n", client.addr, len(s.clients))

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

	if client.name != "" {
		s.broadcast(client, fmt.Sprintf("* %s has left the room\n", client.name))
	}

	log.Printf("connection from %s closed [# connected clients: %d]\n", client.addr, len(s.clients))

	client.conn.Close()
}

func (s *ChatServer) serve(client *client) error {
	defer s.remove(client)

	if _, err := client.conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n")); err != nil {
		return fmt.Errorf("welcome: %w", err)
	}

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		line := string(scanner.Bytes())

		log.Println("received:", line)

		if client.name == "" {
			if err := s.nameClient(client, line); err != nil {
				return err
			}
		} else {
			if err := s.broadcast(client, fmt.Sprintf("[%s] %s\n", client.name, line)); err != nil {
				return fmt.Errorf("broadcast: %w", err)
			}
		}
	}

	return scanner.Err()
}

func (s *ChatServer) nameClient(client *client, name string) error {
	if s.validateClientName(name) {
		client.name = name

		if err := s.broadcast(client, fmt.Sprintf("* %s has entered the room\n", client.name)); err != nil {
			return fmt.Errorf("broadcast: %w", err)
		}

		roomMsg := fmt.Sprintf("* The room contains: %s\n", s.userNamesPresent(client))

		if _, err := client.conn.Write([]byte(roomMsg)); err != nil {
			return fmt.Errorf("room: %w", err)
		}
	} else {
		invalidNameMsg := fmt.Sprintf("invalid name: %s\n", name)
		client.conn.Write([]byte(invalidNameMsg))
		return fmt.Errorf(invalidNameMsg)
	}

	return nil
}

func (s *ChatServer) validateClientName(name string) bool {
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
	for _, client := range s.clients {
		if name == client.name {
			return false
		}
	}

	return true
}

func (s *ChatServer) userNamesPresent(client *client) string {
	var names []string
	for _, c := range s.clients {
		// do not include self or unnamed clients
		if client == c || c.name == "" {
			continue
		}
		names = append(names, c.name)
	}

	return strings.Join(names, ", ")
}

func (s *ChatServer) broadcast(client *client, msg string) error {
	for _, c := range s.clients {
		// do not send to self or unnamed clients
		if client == c || c.name == "" {
			continue
		}

		if _, err := c.conn.Write([]byte(msg)); err != nil {
			return err
		}
	}

	return nil
}
