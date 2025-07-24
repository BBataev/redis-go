package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/entity"
	"github.com/codecrafters-io/redis-starter-go/internal/utils"
)

func HandleCommand(args []string) string {
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
			return "-ERR wrong number of arguments for 'echo' command\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])

	case "SET":
		if len(args) < 3 || len(args) == 4 {
			return "-ERR wrong number of arguments for 'set' command\r\n"
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
		entity.Db[key] = entity.DB{Value: value, TTL: ttl}

		return "+OK\r\n"

	case "GET":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'get' command\r\n"
		}
		key := args[1]
		entry, ok := entity.Db[key]
		if !ok {
			return "$-1\r\n"
		}
		if !entry.TTL.IsZero() && time.Now().After(entry.TTL) {
			delete(entity.Db, key)
			return "$-1\r\n"
		}

		return fmt.Sprintf("$%d\r\n%s\r\n", len(entry.Value.(string)), entry.Value)

	case "DEL":
		if len(args) < 2 {
			return "-ERR wrong number of arguments for 'del' command\r\n"
		}

		count := 0
		for _, key := range args[1:] {
			if _, ok := entity.Db[key]; ok {
				delete(entity.Db, key)
				count++
			}
		}

		return fmt.Sprintf(":%d\r\n", count)

	case "TTL":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'ttl' command\r\n"
		}

		entry, ok := entity.Db[args[1]]
		if !ok {
			return ":-2\r\n"
		}

		if entry.TTL.IsZero() {
			return ":-1\r\n"
		}

		ttl := int(time.Until(entry.TTL).Seconds())
		if ttl < 0 {
			return ":-2\r\n"
		}

		return fmt.Sprintf(":%d\r\n", ttl)

	case "RPUSH":
		if len(args) < 2 {
			return "-ERR wrong number of arguments for 'rpush' command\r\n"
		}

		entry, exists := entity.Db[args[1]]
		list, ok := []string{}, false
		if exists {
			list, ok = entry.Value.([]string)
		}
		if !ok {
			list = []string{}
		}

		list = append(list, args[2:]...)

		entry.Value = list
		entity.Db[args[1]] = entry

		return fmt.Sprintf(":%d\r\n", len(list))

	case "LPUSH":
		if len(args) < 2 {
			return "-ERR wrong number of arguments for 'rpush' command\r\n"
		}

		entry, exists := entity.Db[args[1]]
		list, ok := []string{}, false
		if exists {
			list, ok = entry.Value.([]string)
		}
		if !ok {
			list = []string{}
		}

		list = append(utils.TurnAround(args[2:]), list...)

		entry.Value = list
		entity.Db[args[1]] = entry

		return fmt.Sprintf(":%d\r\n", len(list))

	case "LRANGE":
		if len(args) != 4 {
			return "-ERR wrong number of arguments for 'lrange' command\r\n"
		}

		entry, ok := entity.Db[args[1]]
		if !ok {
			return "*0\r\n"
		}

		slice, ok := entry.Value.([]string)
		if !ok {
			return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		}

		lIndex, err := strconv.Atoi(args[2])
		if err != nil {
			return "-ERR syntax error\r\n"
		}

		rIndex, err := strconv.Atoi(args[3])
		if err != nil {
			return "-ERR syntax error\r\n"
		}

		if lIndex <= -len(slice) {
			lIndex = 0
		}

		if lIndex < 0 {
			lIndex = len(slice) + lIndex
		}

		if rIndex >= len(slice) {
			rIndex = len(slice) - 1
		}

		if rIndex <= -1 {
			rIndex = len(slice) + rIndex
		}

		slice = slice[lIndex : rIndex+1]

		out := fmt.Sprintf("*%d\r\n", len(slice))
		for _, elem := range slice {
			out += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
		}

		return out

	case "LLEN":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'llen' command\r\n"
		}

		entry, ok := entity.Db[args[1]]
		if !ok {
			return ":0\r\n"
		}

		slice, ok := entry.Value.([]string)
		if !ok {
			return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		}

		return fmt.Sprintf(":%d\r\n", len(slice))

	case "LPOP":
		if len(args) > 3 || len(args) < 2 {
			return "-ERR wrong number of arguments for 'lpop' command\r\n"
		}

		entry, ok := entity.Db[args[1]]
		if !ok {
			return "*0\r\n"
		}

		slice, ok := entry.Value.([]string)
		if !ok {
			return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		}

		if len(slice) == 0 {
			return "$-1\r\n"
		}

		n := 1
		if len(args) == 3 {
			var err error
			n, err = strconv.Atoi(args[2])
			if err != nil || n < 0 {
				return "-ERR syntax error\r\n"
			}
		}

		if n > len(slice) {
			n = len(slice)
		}

		left := slice[:n]
		slice = slice[n:]

		entry.Value = slice
		entity.Db[args[1]] = entry

		if len(left) == 1 {
			return fmt.Sprintf("$%d\r\n%s\r\n", len(left[0]), left[0])
		}

		out := fmt.Sprintf("*%d\r\n", len(left))
		for _, elem := range left {
			out += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
		}
		return out

	case "RPOP":
		if len(args) > 3 || len(args) < 2 {
			return "-ERR wrong number of arguments for 'rpop' command\r\n"
		}

		entry, ok := entity.Db[args[1]]
		if !ok {
			return "*0\r\n"
		}

		slice, ok := entry.Value.([]string)
		if !ok {
			return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		}

		if len(slice) == 0 {
			return "$-1\r\n"
		}

		n := 1
		if len(args) == 3 {
			var err error
			n, err = strconv.Atoi(args[2])
			if err != nil || n < 0 {
				return "-ERR syntax error\r\n"
			}
		}

		if n > len(slice) {
			n = len(slice)
		}

		right := slice[len(slice)-n:]
		slice = slice[:len(slice)-n]

		entry.Value = slice
		entity.Db[args[1]] = entry

		if len(right) == 1 {
			return fmt.Sprintf("$%d\r\n%s\r\n", len(right[0]), right[0])
		}

		out := fmt.Sprintf("*%d\r\n", len(right))
		for _, elem := range right {
			out += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
		}
		return out

	default:
		return "-ERR unknown command\r\n"
	}
}
