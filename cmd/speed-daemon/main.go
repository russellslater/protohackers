package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/russellslater/protohackers/cmd/speed-daemon/ticketer"
)

func main() {
	ts := NewTicketServer(5000)
	log.Fatal(ts.Start())
}

type TicketServer struct {
	port     int
	clients  []*client
	listener net.Listener
	sync.Mutex
	ticketManager *ticketer.TicketManager
}

const (
	plateMsg         = 0x20
	wantHeartbeatMsg = 0x40
	iAmCameraMsg     = 0x80
	iAmDispatcherMsg = 0x81

	errorMsg     = 0x10
	ticketMsg    = 0x21
	heartbeatMsg = 0x41
)

type msgType uint8

func NewTicketServer(port int) *TicketServer {
	return &TicketServer{
		port:          port,
		ticketManager: ticketer.NewTicketManager(),
	}
}

func (s *TicketServer) Start() error {
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

func (s *TicketServer) Close() {
	if err := s.listener.Close(); err != nil {
		fmt.Print("error closing connection: %w", err)
	}
}

func (s *TicketServer) connect(conn net.Conn) *client {
	client := &client{
		addr:   conn.RemoteAddr().String(),
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	s.Lock()
	s.clients = append(s.clients, client)
	s.Unlock()

	log.Printf("connection from %s [# connected clients: %d]\n", client.addr, len(s.clients))

	return client
}

func (s *TicketServer) remove(client *client) {
	defer client.conn.Close()

	if client.heartbeatDoneChan != nil {
		client.heartbeatDoneChan <- true
		client.heartbeatTicker.Stop()
	}

	s.Lock()
	for i, c := range s.clients {
		if client == c {
			s.clients[i] = s.clients[len(s.clients)-1]
			s.clients = s.clients[:len(s.clients)-1]
			break
		}
	}
	s.Unlock()

	log.Printf("connection from %s closed [# connected clients: %d]\n", client.addr, len(s.clients))
}

func (s *TicketServer) serve(client *client) error {
	defer s.remove(client)

	log.Printf("incoming connection: %v\n", client.conn)

	for {
		msg, err := client.readMsg()
		if err != nil {
			return err
		}

		switch msg {
		case iAmCameraMsg:
			// echo -e -n '\x80\x00\x7b\x00\x08\x00\x3c' > out
			log.Println("IAmCamera")
			client.camera = &camera{
				road:  client.readUint16(),
				mile:  client.readUint16(),
				limit: client.readUint16(),
			}

			log.Println("Camera: ", client.camera)
		case iAmDispatcherMsg:
			// echo -e -n '\x81\x03\x00\x42\x01\x70\x13\x88' > out
			log.Println("IAmDispatcher")

			client.dispatcher = &dispatcher{
				roads: client.readUint16Array(),
			}

			log.Println("Dispatcher: ", client.dispatcher)
		case plateMsg:
			// echo -e '\x20\x04\x55\x4e\x31\x58\x00\x00\x03\xe8' > out
			log.Println("Plate")

			plate := client.readStr()
			timestamp := client.readUint32()

			log.Println("Plate: ", plate)
			log.Println("Timestamp: ", timestamp)
		case wantHeartbeatMsg:
			// echo -e -n '\x40\x00\x00\x04\xdb' > out
			log.Println("Want Heartbeat")

			interval := client.readUint32()

			client.startHeartbeat(interval)

			log.Println("Interval: ", interval)
		default:
			log.Println("Unknown message")
		}
	}
}
