package main

import (
	"testing"
)

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
