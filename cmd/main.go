package main

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/handler"
	"github.com/codecrafters-io/redis-starter-go/internal/utils"
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

	reader := bufio.NewReader(conn)
	for {
		val, err := utils.RESParser(reader)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Клиент закрыл соединение")
				return
			}
			fmt.Printf("Ошибка: %v, val: %#v\n", err, val)
			continue
		}

		args := val.([]string)
		res := handler.HandleCommand(args)

		_, _ = conn.Write([]byte(res))
	}
}
