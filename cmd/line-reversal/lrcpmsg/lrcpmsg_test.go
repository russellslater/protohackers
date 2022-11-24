package lrcpmsg_test

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers/cmd/line-reversal/lrcpmsg"
)

func TestParseMsg(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		input []byte
		want  interface{}
		err   error
	}{
		{
			"Input is nil",
			nil,
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is an empty byte array",
			[]byte{},
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a byte array over max size",
			make([]byte, 1000),
			nil,
			fmt.Errorf("message exceeds max size"),
		},
		{
			"Input doesn't start or end with slash but valid otherwise",
			[]byte("connect/123"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input doesn't start with slash but valid otherwise",
			[]byte("connect/123/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input doesn't end with slash but valid otherwise",
			[]byte("/connect/123"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a valid connect message",
			[]byte("/connect/123/"),
			lrcpmsg.ConnectMsg{123},
			nil,
		},
		{
			"Input is a connect message with the largest possible session ID",
			[]byte("/connect/2147483647/"),
			lrcpmsg.ConnectMsg{2147483647},
			nil,
		},
		{
			"Input is a connect message with a session ID of 0",
			[]byte("/connect/0/"),
			lrcpmsg.ConnectMsg{0},
			nil,
		},
		{
			"Input is a connect message with an invalid session ID",
			[]byte("/connect/1a2b3c/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a connect message with a session ID larger than the permitted max",
			[]byte("/connect/2147483648/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is a connect message with a negative integer for a session ID",
			[]byte("/connect/-1000/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a valid close message",
			[]byte("/close/123/"),
			lrcpmsg.CloseMsg{123},
			nil,
		},
		{
			"Input is a close message with the largest possible session ID",
			[]byte("/close/2147483647/"),
			lrcpmsg.CloseMsg{2147483647},
			nil,
		},
		{
			"Input is a close message with a session ID of 0",
			[]byte("/close/0/"),
			lrcpmsg.CloseMsg{0},
			nil,
		},
		{
			"Input is a close message with an invalid session ID",
			[]byte("/close/1a2b3c/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a close message with a session ID larger than the permitted max",
			[]byte("/close/2147483648/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is a close message with a negative integer for a session ID",
			[]byte("/close/-1000/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a valid ack message",
			[]byte("/ack/98756/104/"),
			lrcpmsg.AckMsg{98756, 104},
			nil,
		},
		{
			"Input is an ack message with the largest possible session ID and length",
			[]byte("/ack/2147483647/2147483647/"),
			lrcpmsg.AckMsg{2147483647, 2147483647},
			nil,
		},
		{
			"Input is an ack message with a session ID and length of 0",
			[]byte("/ack/0/0/"),
			lrcpmsg.AckMsg{0, 0},
			nil,
		},
		{
			"Input is an ack message with an invalid session ID",
			[]byte("/ack/1a2b3c/1234/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is an ack message with an invalid length",
			[]byte("/ack/1234/1a2b3c/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is an ack message with a session ID larger than the permitted max",
			[]byte("/ack/2147483648/1234/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is an ack message with a length larger than the permitted max",
			[]byte("/ack/1234/2147483648/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is an ack message with a negative integer for a session ID",
			[]byte("/ack/-1000/1234"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is an ack message with a negative integer for a length",
			[]byte("/ack/1234/-1000"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a valid ack message",
			[]byte("/data/12345/0/Hello, world!\n/"),
			lrcpmsg.DataMsg{12345, 0, []byte("Hello, world!\n")},
			nil,
		},
		{
			"Input is a data message with the largest possible session ID and pos",
			[]byte("/data/2147483647/2147483647/abcdefg/"),
			lrcpmsg.DataMsg{2147483647, 2147483647, []byte("abcdefg")},
			nil,
		},
		{
			"Input is a data message with empty data",
			[]byte("/data/1234/10//"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a data message with a session ID and pos of 0",
			[]byte(`/data/0/0/foo\/bar\\baz/`),
			lrcpmsg.DataMsg{0, 0, []byte(`foo\/bar\\baz`)},
			nil,
		},
		{
			"Input is a data message with an invalid session ID",
			[]byte("/data/1a2b3c/1234/Hello/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a data message with an invalid pos",
			[]byte("/data/1234/1a2b3c/Hello/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a data message with a session ID larger than the permitted max",
			[]byte("/data/2147483648/1234/Hello, World!/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is a data message with a pos larger than the permitted max",
			[]byte("/data/1234/2147483648/Hello, World!/"),
			nil,
			fmt.Errorf("numeric field exceeds maximum numeric value"),
		},
		{
			"Input is a data message with a negative integer for a session ID",
			[]byte("/data/-1000/1234/Hello, World!/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
		{
			"Input is a data message with a negative integer for a pos",
			[]byte("/data/1234/-1000/Hello, World!/"),
			nil,
			fmt.Errorf("invalid message format"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			got, err := lrcpmsg.ParseMsg(tc.input)

			is.Equal(got, tc.want)
			is.Equal(err, tc.err)
		})
	}
}
