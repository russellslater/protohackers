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
			fmt.Errorf("session ID exceeds maximum numeric value"),
		},
		{
			"Input is a connect message with a negative integer for a session ID",
			[]byte("/connect/-1000/"),
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
