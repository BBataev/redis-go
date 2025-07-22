package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	fmt.Println("Listening on :6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewScanner(conn)
	for reader.Scan() {
		text := reader.Text()
		if strings.TrimSpace(text) == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		}
	}
}
