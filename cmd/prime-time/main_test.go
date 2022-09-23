package main

import (
	"bufio"
	"net"
	"testing"

	"github.com/matryer/is"
)

func TestPrimeTimeHandler(t *testing.T) {
	is := is.New(t)

	client, server := net.Pipe()

	go handle(server)

	clientScanner := bufio.NewScanner(client)

	client.Write([]byte("{\"method\":\"isPrime\",\"number\":1}\n"))
	clientScanner.Scan()
	is.Equal(clientScanner.Text(), "{\"method\":\"isPrime\",\"prime\":false}")

	client.Write([]byte("{\"method\":\"isPrime\",\"number\":7}\n"))
	clientScanner.Scan()
	is.Equal(clientScanner.Text(), "{\"method\":\"isPrime\",\"prime\":true}")

	client.Write([]byte("{\"method\":\"isPrime\",\"number\":ABC}\n"))
	clientScanner.Scan()
	is.Equal(clientScanner.Text(), "invalid request")

	client.Write([]byte("{\"method\":\"isPrime\",\"number\":7}\n"))
	clientScanner.Scan()
	is.Equal(len(clientScanner.Text()), 0) // should be disconnected
}

func TestValidLines(t *testing.T) {
	tt := []struct {
		line      []byte
		wantBytes []byte
		wantValid bool
	}{
		{
			line:      []byte("{\"method\":\"isPrime\",\"number\":1}"),
			wantBytes: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
			wantValid: true,
		},
		{
			line:      []byte("{\"method\":\"isPrime\",\"number\":2}"),
			wantBytes: []byte("{\"method\":\"isPrime\",\"prime\":true}\n"),
			wantValid: true,
		},
		{
			line:      []byte("{\"method\":\"isPrime\",\"number\":999983}"),
			wantBytes: []byte("{\"method\":\"isPrime\",\"prime\":true}\n"),
			wantValid: true,
		},
		{
			line:      []byte("{\"method\":\"isPrime\",\"number\":123,\"ignoreProperty\":true}"),
			wantBytes: []byte("{\"method\":\"isPrime\",\"prime\":false}\n"),
			wantValid: true,
		},
		{
			line:      []byte("{\"method\":\"isPrime\",\"number\":\"not-a-number\"}"),
			wantBytes: []byte("invalid request\n"),
			wantValid: false,
		},
		{
			line:      []byte("{}"),
			wantBytes: []byte("invalid request\n"),
			wantValid: false,
		},
		{
			line:      []byte("garbage_request"),
			wantBytes: []byte("invalid request\n"),
			wantValid: false,
		},
	}
	for _, tc := range tt {
		gotBytes, gotValid, _ := handleLine(tc.line)
		if string(gotBytes) != string(tc.wantBytes) {
			t.Errorf("got bytes %q, want %q", gotBytes, tc.wantBytes)
		}
		if gotValid != tc.wantValid {
			t.Errorf("got valid %t, want %t", gotValid, tc.wantValid)
		}
	}
}
