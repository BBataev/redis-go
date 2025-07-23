package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

var db = make(map[string]string)

func main() {
	l, err := net.Listen("tcp", ":6378")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	fmt.Println("Listening on :6378")

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
		val, err := RESParser(reader)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Клиент закрыл соединение")
				return
			}
			fmt.Printf("Ошибка: %v, val: %#v\n", err, val)
			continue
		}

		args := val.([]string)
		res := handleCommand(args)

		_, _ = conn.Write([]byte(res))
	}
}

func RESParser(r *bufio.Reader) (interface{}, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if prefix != '*' {
		return nil, fmt.Errorf("expected * symbol, got %q", prefix)
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unexpecred behaviour, got %s", err)
	}

	count, err := strconv.Atoi(strings.TrimSuffix(line, "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("unexpecred behaviour, got %s", err)
	}

	args := make([]string, count)

	for i := 0; i < count; i++ {
		prefix, err = r.ReadByte()
		if err != nil {
			return nil, err
		}

		if prefix != '$' {
			return nil, fmt.Errorf("expected $ symbol, got %q", prefix)
		}

		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("unexpecred behaviour, got %s", err)
		}

		countBulk, err := strconv.Atoi(strings.TrimSuffix(line, "\r\n"))
		if err != nil {
			return nil, fmt.Errorf("unexpecred behaviour, got %s", err)
		}

		buf := make([]byte, countBulk)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string content: %w", err)
		}

		crlf := make([]byte, 2)
		_, err = io.ReadFull(r, crlf)
		if err != nil {
			return nil, fmt.Errorf("failed to read CRLF: %w", err)
		}

		args[i] = string(buf)
	}

	return args, nil
}

func handleCommand(args []string) string {
	if len(args) == 0 {
		return "-ERR empty command\r\n"
	}

	switch strings.ToUpper(args[0]) {
	case "PING":
		if len(args) == 1 {
			return "+PONG\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
	case "ECHO":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'ECHO' command\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
	case "SET":
		if len(args) != 3 {
			return "-ERR wrong number of arguments for 'SET' command\r\n"
		}
		key := args[1]
		value := args[2]
		db[key] = value
		return "+OK\r\n"
	case "GET":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'GET' command\r\n"
		}
		key := args[1]
		value, ok := db[key]
		if !ok {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	default:
		return "-ERR unknown command\r\n"
	}
}
