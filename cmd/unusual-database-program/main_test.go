package main

import (
	"net"
	"testing"

	"github.com/matryer/is"
)

func init() {
	s := NewUnusualDatabaseServer(5000, "")
	go s.Start()
}

func TestUnusualDatabaseServer(t *testing.T) {
	type request struct {
		payload          []byte
		expectedResponse []byte
	}

	tt := []struct {
		name     string
		requests []request
	}{
		{
			name: "Version Not Overridden",
			requests: []request{
				{payload: []byte("version"), expectedResponse: []byte("version=Ken's Key-Value Store 1.0")},
				{payload: []byte("version=1.0.1"), expectedResponse: nil},
				{payload: []byte("version"), expectedResponse: []byte("version=Ken's Key-Value Store 1.0")},
			},
		},
		{
			name: "Foo Bars",
			requests: []request{
				{payload: []byte("foo=bar"), expectedResponse: nil},
				{payload: []byte("foo"), expectedResponse: []byte("foo=bar")},
				{payload: []byte("foo=bar=baz"), expectedResponse: nil},
				{payload: []byte("foo"), expectedResponse: []byte("foo=bar=baz")},
				{payload: []byte("foo="), expectedResponse: nil},
				{payload: []byte("foo"), expectedResponse: []byte("foo=")},
				{payload: []byte("foo==="), expectedResponse: nil},
				{payload: []byte("foo"), expectedResponse: []byte("foo===")},
				{payload: []byte("=foo"), expectedResponse: nil},
				{payload: []byte(""), expectedResponse: []byte("=foo")},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			addr, _ := net.ResolveUDPAddr("udp", ":5000")
			conn, err := net.DialUDP("udp", nil, addr)

			if err != nil {
				t.Errorf("could not connect to server: %v", err)
			}

			defer conn.Close()

			for _, request := range tc.requests {
				if _, err := conn.Write(request.payload); err != nil {
					t.Errorf("could not write payload to server: %v", err)
				}

				// write requests don't expect responses
				if request.expectedResponse == nil {
					continue
				}

				got := make([]byte, 1000)
				if n, err := conn.Read(got); err == nil {
					is.Equal(string(got[:n]), string(request.expectedResponse)) // response did not match
				} else {
					t.Errorf("could not read from connection: %v", err)
				}
			}
		})
	}
}
