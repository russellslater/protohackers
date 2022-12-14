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
				log.Println(err.Error())
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

	if client.isHeartbeatEnabled() {
		client.stopHeartbeat()
	}

	s.ticketManager.RemoveDispatcher(client)

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
			if client.isIdentified() {
				client.sendError("Already identified")
				return nil // disconnect gracefully
			}

			client.camera = &camera{
				road:  client.readUint16(),
				mile:  client.readUint16(),
				limit: client.readUint16(),
			}

			log.Printf("Camera: %v\n", client.camera)
		case iAmDispatcherMsg:
			if client.isIdentified() {
				client.sendError("Already identified")
				return nil // disconnect gracefully
			}

			client.dispatcher = &dispatcher{roads: client.readUint16Array()}

			log.Printf("Dispatcher: %v\n", client.dispatcher)

			s.ticketManager.AddDispatcher(client)
		case plateMsg:
			if !client.isCamera() {
				client.sendError("Client must identify as camera to observe plate")
				return nil // disconnect gracefully
			}

			plate := client.readStr()
			timestamp := client.readUint32()

			ob := &ticketer.Observation{
				Road: &ticketer.Road{
					ID:    ticketer.RoadID(client.camera.road),
					Limit: client.camera.limit,
				},
				Mile:      client.camera.mile,
				Plate:     plate,
				Timestamp: timestamp,
			}

			log.Printf("Observation: %v\n", ob)

			s.ticketManager.Observe(ob)
		case wantHeartbeatMsg:
			if client.isHeartbeatEnabled() {
				client.sendError("Heartbeat already enabled")
				return nil // disconnect gracefully
			}

			interval := client.readUint32()

			client.startHeartbeat(interval)
		default:
			client.sendError("Unknown message")
			return nil // disconnect gracefully
		}
	}
}
