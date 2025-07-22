package main

import (
	"net"
	"os"
)

var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		panic(err)
	}
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("+PONG\r\n"))
}
