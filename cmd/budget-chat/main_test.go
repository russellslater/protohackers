package main

import (
	"bufio"
	"net"
	"sync"
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

type messageAssertion struct {
	scanner  *bufio.Scanner
	expected string
}

func (m *messageExpecter) waitAssert(msgAssertions []messageAssertion) {
	var wg sync.WaitGroup
	wg.Add(len(msgAssertions))

	for _, msgAssertion := range msgAssertions {
		go func(msgAssertion messageAssertion) {
			defer wg.Done()
			m.assert(msgAssertion.scanner, msgAssertion.expected)
		}(msgAssertion)
	}

	wg.Wait()
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

	m.waitAssert([]messageAssertion{
		{scanner: bobClientScanner, expected: "* The room contains: Alice"},
		{scanner: aliceClientScanner, expected: "* Bob has entered the room"},
	})

	// Chieko joins the chat
	m.assert(chiekoClientScanner, "Welcome to budgetchat! What shall I call you?")

	chiekoClientConn.Write([]byte("Chieko\n"))

	m.waitAssert([]messageAssertion{
		{scanner: chiekoClientScanner, expected: "* The room contains: Alice, Bob"},
		{scanner: aliceClientScanner, expected: "* Chieko has entered the room"},
		{scanner: bobClientScanner, expected: "* Chieko has entered the room"},
	})

	// Bob talks
	bobClientConn.Write([]byte("Hey, Alice!\n"))

	m.waitAssert([]messageAssertion{
		{scanner: aliceClientScanner, expected: "[Bob] Hey, Alice!"},
		{scanner: chiekoClientScanner, expected: "[Bob] Hey, Alice!"},
	})

	// Alice talks
	aliceClientConn.Write([]byte("Bob, is that you?\n"))

	m.waitAssert([]messageAssertion{
		{scanner: bobClientScanner, expected: "[Alice] Bob, is that you?"},
		{scanner: chiekoClientScanner, expected: "[Alice] Bob, is that you?"},
	})

	// Alice talks again
	aliceClientConn.Write([]byte("Talk to me, Bob!\n"))

	m.waitAssert([]messageAssertion{
		{scanner: bobClientScanner, expected: "[Alice] Talk to me, Bob!"},
		{scanner: chiekoClientScanner, expected: "[Alice] Talk to me, Bob!"},
	})

	// Chieko talks
	chiekoClientConn.Write([]byte("Alice, wait! I'll talk to you!\n"))

	m.waitAssert([]messageAssertion{
		{scanner: aliceClientScanner, expected: "[Chieko] Alice, wait! I'll talk to you!"},
		{scanner: bobClientScanner, expected: "[Chieko] Alice, wait! I'll talk to you!"},
	})

	// Alice leaves
	aliceClientConn.Close()

	m.waitAssert([]messageAssertion{
		{scanner: chiekoClientScanner, expected: "* Alice has left the room"},
		{scanner: bobClientScanner, expected: "* Alice has left the room"},
	})

	// Bob leaves
	bobClientConn.Close()

	m.waitAssert([]messageAssertion{
		{scanner: chiekoClientScanner, expected: "* Bob has left the room"},
	})

	// Chieko leaves
	chiekoClientConn.Close()
}
