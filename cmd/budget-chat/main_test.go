package main

import (
	"bufio"
	"net"
	"testing"

	"github.com/matryer/is"
)

type messageExpecter struct {
	is *is.I
}

func newMessageExpecter(t *testing.T) *messageExpecter {
	return &messageExpecter{is: is.New(t)}
}

func (m *messageExpecter) assert(scn *bufio.Scanner, expected string) {
	scn.Scan()
	m.is.Equal(scn.Text(), expected)
}

func TestBudgetChatScenario(t *testing.T) {
	s := NewChatServer(5000)

	aliceClientConn, aliceServerConn := net.Pipe()
	bobClientConn, bobServerConn := net.Pipe()
	chiekoClientConn, chiekoServerConn := net.Pipe()

	aliceClient := s.connect(aliceServerConn)
	bobClient := s.connect(bobServerConn)
	chiekoClient := s.connect(chiekoServerConn)

	go func() {
		s.serve(aliceClient)
	}()

	go func() {
		s.serve(bobClient)
	}()

	go func() {
		s.serve(chiekoClient)
	}()

	aliceClientScanner := bufio.NewScanner(aliceClientConn)
	bobClientScanner := bufio.NewScanner(bobClientConn)
	chiekoClientScanner := bufio.NewScanner(chiekoClientConn)

	m := newMessageExpecter(t)

	// Alice joins the chat
	m.assert(aliceClientScanner, "Welcome to budgetchat! What shall I call you?")

	aliceClientConn.Write([]byte("Alice\n"))

	m.assert(aliceClientScanner, "* The room contains: ")

	// Bob joins the chat
	m.assert(bobClientScanner, "Welcome to budgetchat! What shall I call you?")

	bobClientConn.Write([]byte("Bob\n"))

	m.assert(aliceClientScanner, "* Bob has entered the room")
	m.assert(bobClientScanner, "* The room contains: Alice")

	// Chieko joins the chat
	m.assert(chiekoClientScanner, "Welcome to budgetchat! What shall I call you?")

	chiekoClientConn.Write([]byte("Chieko\n"))

	m.assert(aliceClientScanner, "* Chieko has entered the room")
	m.assert(bobClientScanner, "* Chieko has entered the room")
	m.assert(chiekoClientScanner, "* The room contains: Alice, Bob")

	// Bob talks
	bobClientConn.Write([]byte("Hey, Alice!\n"))

	m.assert(aliceClientScanner, "[Bob] Hey, Alice!")
	m.assert(chiekoClientScanner, "[Bob] Hey, Alice!")

	// Alice talks
	aliceClientConn.Write([]byte("Bob, is that you?\n"))

	m.assert(bobClientScanner, "[Alice] Bob, is that you?")
	m.assert(chiekoClientScanner, "[Alice] Bob, is that you?")

	// Alice talks again
	aliceClientConn.Write([]byte("Talk to me, Bob!\n"))

	m.assert(bobClientScanner, "[Alice] Talk to me, Bob!")
	m.assert(chiekoClientScanner, "[Alice] Talk to me, Bob!")

	// Chieko talks
	chiekoClientConn.Write([]byte("Alice, wait! I'll talk to you!\n"))

	m.assert(aliceClientScanner, "[Chieko] Alice, wait! I'll talk to you!")
	m.assert(bobClientScanner, "[Chieko] Alice, wait! I'll talk to you!")

	// Alice leaves
	aliceClientConn.Close()

	m.assert(chiekoClientScanner, "* Alice has left the room")
	m.assert(bobClientScanner, "* Alice has left the room")

	// Bob leaves, Chieko leaves
	bobClientConn.Close()
	chiekoClientConn.Close()
}
