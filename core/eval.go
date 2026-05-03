package core

import (
	"errors"
	"io"
	"strconv"
	"time"
)

func EvalAndRespond(cmd *RedisCommand, conn io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, conn)
	case "SET":
		return evalSet(cmd.Args, conn)
	case "GET":
		return evalGet(cmd.Args, conn)
	case "TTL":
		return evalTTL(cmd.Args, conn)
	default:
		return errors.New("ERR unknown command '" + cmd.Cmd + "'")
	}
}

func evalPing(args []string, conn io.ReadWriter) error {
	var b []byte

	if len(args) > 1 {
		return errors.New("ERR wrong number of arguments passed for 'ping' commnad")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := conn.Write(b)
	return err
}

var RESP_NIL []byte = []byte("$-1\r\n")

func evalSet(args []string, conn io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	var key, value string
	var exDurationMs int64 = -1

	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("ERR syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return errors.New("ERR value is not an integer or out of range")
			}

			exDurationMs = exDurationSec * 1000
		default:
			return errors.New("ERR syntax error")
		}
	}

	Put(key, NewObj(value, exDurationMs))
	conn.Write(Encode("OK", true))
	return nil
}

func evalGet(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}

	var key string = args[0]

	obj := Get(key)

	if obj == nil {
		conn.Write(RESP_NIL)
		return nil
	}

	if obj.expiresAt != -1 && obj.expiresAt <= time.Now().UnixMilli() {
		conn.Write(RESP_NIL)
		return nil
	}

	conn.Write(Encode(obj.value, false))
	return nil
}

func evalTTL(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		conn.Write(Encode(int64(-2), false))
		return nil
	}

	if obj.expiresAt == -1 {
		conn.Write(Encode(int64(-1), false))
		return nil
	}

	durationMs := obj.expiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		conn.Write(Encode(int64(-2), false))
		return nil
	}

	conn.Write(Encode(int64(durationMs/1000), false))
	return nil
}
