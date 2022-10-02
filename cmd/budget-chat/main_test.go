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
	t.Parallel()

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

func TestNameValidation(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Alice", input: "Alice", expected: "* The room contains: "},
		{name: "Bob", input: "Bob", expected: "* The room contains: "},
		{name: "Chieko", input: "Chieko", expected: "* The room contains: "},
		{name: "All Numbers", input: "0123456789", expected: "* The room contains: "},
		{name: "Non-Alphanumeric Character", input: "Alice!", expected: "invalid name: Alice!"},
		{name: "Empty String", input: "", expected: "invalid name: "},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newMessageExpecter(t)

			s := NewChatServer(5000)

			clientConn, serverConn := net.Pipe()

			client := s.connect(serverConn)

			go func() {
				s.serve(client)
			}()

			clientScanner := bufio.NewScanner(clientConn)

			m.assert(clientScanner, "Welcome to budgetchat! What shall I call you?")

			clientConn.Write([]byte(tc.input + "\n"))

			m.assert(clientScanner, tc.expected)

			clientConn.Close()
		})
	}
}

func TestDuplicateName(t *testing.T) {
	t.Parallel()
	m := newMessageExpecter(t)

	s := NewChatServer(5000)

	uniqueClientConn, uniqueServerConn := net.Pipe()
	dupeClientConn, dupeServerConn := net.Pipe()

	uniqueClient := s.connect(uniqueServerConn)
	dupeClient := s.connect(dupeServerConn)

	go func() {
		s.serve(uniqueClient)
	}()

	go func() {
		s.serve(dupeClient)
	}()

	uniqueClientScanner := bufio.NewScanner(uniqueClientConn)
	client2Scanner := bufio.NewScanner(dupeClientConn)

	m.assert(uniqueClientScanner, "Welcome to budgetchat! What shall I call you?")

	uniqueClientConn.Write([]byte("Uniquename\n"))

	m.assert(uniqueClientScanner, "* The room contains: ")

	m.assert(client2Scanner, "Welcome to budgetchat! What shall I call you?")

	// Duplicate name!
	dupeClientConn.Write([]byte("Uniquename\n"))

	m.assert(client2Scanner, "invalid name: Uniquename")

	uniqueClientConn.Close()
	dupeClientConn.Close()
}
