package main

import (
	"bufio"
	"net"
	"testing"

	"github.com/matryer/is"
)

func TestHandle(t *testing.T) {
	is := is.New(t)

	client, server := net.Pipe()

	go echo(server)

	clientScanner := bufio.NewScanner(client)

	client.Write([]byte("echo\n"))
	clientScanner.Scan()
	is.Equal(clientScanner.Text(), "echo")

	client.Write([]byte("hello world\n"))
	clientScanner.Scan()
	is.Equal(clientScanner.Text(), "hello world")

	server.Close()

	client.Write([]byte("hello?\n"))
	clientScanner.Scan()
	is.Equal(len(clientScanner.Text()), 0) // server hung up

	client.Close()
}
