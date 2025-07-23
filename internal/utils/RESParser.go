package utils

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
