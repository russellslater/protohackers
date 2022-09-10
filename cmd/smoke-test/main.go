package main

import (
	"io"
	"net"

	"github.com/russellslater/protohackers"
)

func main() {
	protohackers.ListenAndAccept(5000, echo)
}

func echo(conn net.Conn) error {
	defer conn.Close()

	_, err := io.Copy(conn, conn)
	return err
}
