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
	closeMsgRegex, _   = regexp.Compile(`^/close/(?P<SessionID>\d{1,10})/$`)
	ackMsgRegex, _     = regexp.Compile(`^/ack/(?P<SessionID>\d{1,10})/(?P<Length>\d{1,10})/$`)
	dataMsgRegex, _    = regexp.Compile(`^\/data\/(?P<SessionID>\d{1,10})\/(?P<Pos>\d{1,10})\/(?s)(?P<Data>.+)\/$`)
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
		if len(regexResult) == 2 {
			sessionID, err = parseNumericField(string(regexResult[1]))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("invalid message format")
		}

		return ConnectMsg{sessionID}, nil
	case bytes.HasPrefix(msg, []byte("/close")):
		sessionID := 0
		var err error

		regexResult := closeMsgRegex.FindSubmatch(msg)
		if len(regexResult) == 2 {
			sessionID, err = parseNumericField(string(regexResult[1]))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("invalid message format")
		}

		return CloseMsg{sessionID}, nil
	case bytes.HasPrefix(msg, []byte("/ack")):
		sessionID, length := 0, 0
		var err error

		regexResult := ackMsgRegex.FindSubmatch(msg)

		if len(regexResult) == 3 {
			sessionID, err = parseNumericField(string(regexResult[1]))
			if err != nil {
				return nil, err
			}
			length, err = parseNumericField(string(regexResult[2]))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("invalid message format")
		}

		return AckMsg{sessionID, length}, nil
	case bytes.HasPrefix(msg, []byte("/data")):
		sessionID, pos := 0, 0
		var data []byte
		var err error

		regexResult := dataMsgRegex.FindSubmatch(msg)

		if len(regexResult) == 4 {
			sessionID, err = parseNumericField(string(regexResult[1]))
			if err != nil {
				return nil, err
			}
			pos, err = parseNumericField(string(regexResult[2]))
			if err != nil {
				return nil, err
			}
			data = regexResult[3]
		} else {
			return nil, fmt.Errorf("invalid message format")
		}

		return DataMsg{sessionID, pos, data}, nil
	}

	return nil, nil
}

func parseNumericField(numStr string) (int, error) {
	num, err := strconv.Atoi(numStr)

	if err != nil {
		return 0, fmt.Errorf("invalid message format")
	}

	if num > math.MaxInt32 {
		return 0, fmt.Errorf("numeric field exceeds maximum numeric value")
	}

	return num, nil
}
