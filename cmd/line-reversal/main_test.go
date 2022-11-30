package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/matryer/is"
)

func init() {
	s := NewLineReversalServer(5000, "")
	go s.Start()
}

func waitForServerStart(t *testing.T) {
	isStarted := false
	rand.Seed(time.Now().UnixNano())

	addr, _ := net.ResolveUDPAddr("udp", ":5000")

	for i := 0; i < 10; i++ {
		conn, _ := net.DialUDP("udp", nil, addr)
		conn.Write([]byte("/connect/0/"))

		_, err := conn.Read(make([]byte, 1000))
		if err == nil {
			isStarted = true
			break
		}

		conn.Close()

		t := math.Pow(2, float64(i))*1000 + (rand.Float64() * 1000)
		backoff := math.Min(t, 10000)
		time.Sleep(time.Millisecond * time.Duration(backoff))
	}

	if !isStarted {
		t.Error("could not connect to server")
	}
}

func TestLineReversalServer(t *testing.T) {
	waitForServerStart(t)

	type request struct {
		payload          []byte
		expectedResponse []byte
	}

	tt := []struct {
		name     string
		requests []request
	}{
		{
			name: "Protohackers example scenario",
			requests: []request{
				{payload: []byte("/connect/12345/"), expectedResponse: []byte("/ack/12345/0/")},
				{payload: []byte(`/data/12345/0/hello\\n/`), expectedResponse: []byte("/ack/12345/6/")},
				{payload: nil, expectedResponse: []byte(`/data/12345/0/olleh\\n/`)},
				{payload: []byte("/ack/12345/6/"), expectedResponse: nil},
				{payload: []byte(`/data/12345/6/Hello, world!\\n/`), expectedResponse: []byte("/ack/12345/20/")},
				{payload: nil, expectedResponse: []byte(`/data/12345/6/!dlrow ,olleH\\n/`)},
				{payload: []byte("/ack/12345/20/"), expectedResponse: nil},
				{payload: []byte("/close/12345/"), expectedResponse: []byte("/close/12345/")},
			},
		},
		{
			name: "Multiple connects",
			requests: []request{
				{payload: []byte("/connect/987654/"), expectedResponse: []byte("/ack/987654/0/")},
				{payload: []byte("/connect/987654/"), expectedResponse: []byte("/ack/987654/0/")},
				{payload: []byte(`/data/987654/0/hello/`), expectedResponse: []byte("/ack/987654/5/")},
				{payload: []byte("/connect/987654/"), expectedResponse: []byte("/ack/987654/0/")},
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
				if request.payload != nil {
					if _, err := conn.Write(request.payload); err != nil {
						t.Errorf("could not write payload to server: %v", err)
					}
				}

				// response not expected
				if request.expectedResponse == nil {
					continue
				}

				fmt.Println(request)

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

func TestSessionAppendData(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		data []string
		want []int
	}{
		{
			"Empty string",
			[]string{""},
			[]int{0},
		},
		{
			"Hello, world",
			[]string{"Hello,", " world!", "\n"},
			[]int{6, 13, 14},
		},
		{
			"Single new line character",
			[]string{"\n", "", "\n", "", "Foo", "Bar", "", "\n"},
			[]int{1, 1, 2, 2, 5, 8, 8, 9},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			s := NewSession(0, nil)

			var got int
			for i, d := range tc.data {
				got = s.AppendData(d)
				is.Equal(got, tc.want[i])
			}

			is.Equal(s.ReceivedPos, got)
		})
	}
}

func TestSessionCompletedLinesOutOfBounds(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	s := NewSession(0, nil)

	gotStrs, gotLen := s.CompletedLines(100) // out of bounds

	is.Equal(gotStrs, nil)
	is.Equal(gotLen, 0)
}

func TestSessionCompletedLines(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		data     string
		pos      int
		wantStrs []string
		wantLen  int
	}{
		{
			"Incomplete line",
			"Hello, world!",
			0,
			nil,
			0,
		},
		{
			"Single line",
			"Hello, world!\n",
			0,
			[]string{"Hello, world!"},
			14,
		},
		{
			"Two lines",
			"Hello, world!\nHere's another line.\n",
			0,
			[]string{"Hello, world!", "Here's another line."},
			35,
		},
		{
			"Empty string",
			"",
			0,
			nil,
			0,
		},
		{
			"Start part way through second line",
			"Apple\nOrange\nPear\n",
			8,
			[]string{"ange", "Pear"},
			10,
		},
		{
			"Start part way through third line",
			"Apple\nOrange\nPear\n",
			15,
			[]string{"ar"},
			3,
		},
		{
			"Four empty lines",
			"\n\n\n\n",
			0,
			[]string{"", "", "", ""},
			4,
		},
		{
			"Last two of four empty lines",
			"\n\n\n\n",
			2,
			[]string{"", ""},
			2,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			s := NewSession(0, nil)
			s.AppendData(tc.data)

			gotStrs, gotLen := s.CompletedLines(tc.pos)

			is.Equal(gotStrs, tc.wantStrs)
			is.Equal(gotLen, tc.wantLen)
		})
	}
}
