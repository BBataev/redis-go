package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type DB struct {
	Value string
	TTL   time.Time
}

var db = make(map[string]DB)

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
		if len(args) < 3 {
			return "-ERR wrong number of arguments for 'SET' command\r\n"
		}
		key := args[1]
		value := args[2]

		var ttl time.Time
		if len(args) == 5 && strings.ToUpper(args[3]) == "PX" {
			ms, err := strconv.Atoi(args[4])
			if err != nil {
				return "-ERR PX argument should be a number"
			}
			ttl = time.Now().Add(time.Duration(ms) * time.Millisecond)
		}
		db[key] = DB{Value: value, TTL: ttl}

		return "+OK\r\n"
	case "GET":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'GET' command\r\n"
		}
		key := args[1]
		entry, ok := db[key]
		if !ok {
			return "$-1\r\n"
		}
		if !entry.TTL.IsZero() && time.Now().After(entry.TTL) {
			delete(db, key)
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(entry.Value), entry.Value)
	default:
		return "-ERR unknown command\r\n"
	}
}
