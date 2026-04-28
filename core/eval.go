package core

import (
	"errors"
	"io"
)

func EvalAndRespond(cmd *RedisCommand, conn io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, conn)
	case "COMMAND":
		conn.Write(Encode("OK", true))
		return nil
	case "CONFIG":
		conn.Write(Encode("OK", true))
		return nil
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
