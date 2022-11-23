package lrcpmsg

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var (
	connectMsgRegex, _ = regexp.Compile(`^/connect/(?P<SessionID>\d{1,10})/$`)
)

type ConnectMsg struct {
	SessionID int
}

type DataMsg struct {
	SessionID int
	Pos       int
	Data      []byte
}

type AckMsg struct {
	SessionID int
	Length    int
}

type CloseMsg struct {
	SessionID int
}

// Packet contents must begin with a forward slash, end with a forward slash,
// have a valid message type, and have the correct number of fields for the message type.
// Numeric field values must be smaller than 2147483648.
// LRCP messages must be smaller than 1000 bytes.
func ParseMsg(msg []byte) (interface{}, error) {
	if len(msg) >= 1000 {
		return nil, fmt.Errorf("message exceeds max size")
	}

	slash := []byte("/")

	if !bytes.HasPrefix(msg, slash) || !bytes.HasSuffix(msg, slash) {
		return nil, fmt.Errorf("invalid message format")
	}

	switch {
	case bytes.HasPrefix(msg, []byte("/connect")):
		sessionID := 0
		var err error

		regexResult := connectMsgRegex.FindSubmatch(msg)
		if len(regexResult) > 1 {
			sessionID, err = strconv.Atoi(string(regexResult[1]))

			if err != nil {
				return nil, fmt.Errorf("invalid message format")
			}

			if sessionID > math.MaxInt32 {
				return nil, fmt.Errorf("session ID exceeds maximum numeric value")
			}
		} else {
			return nil, fmt.Errorf("invalid message format")
		}

		return ConnectMsg{sessionID}, nil
	}

	return nil, nil
}
