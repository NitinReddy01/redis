package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/NitinReddy01/redis/config"
	"github.com/NitinReddy01/redis/core"
)

func RunSyncTCPServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.PORT))
	if err != nil {
		log.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()
	log.Printf("server listening on port:%d", config.PORT)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
		}
		log.Printf("client connected: %s", conn.RemoteAddr().String())

		for {
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()
				log.Printf("client disconnected: %s", conn.RemoteAddr().String())
				if err == io.EOF {
					break
				}
				log.Print(err)
			}
			respond(cmd, conn)
		}
	}
}

func readCommand(conn io.ReadWriter) (*core.RedisCommand, error) {
	var buf []byte = make([]byte, 512)

	n, err := conn.Read(buf)

	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])

	if err != nil {
		return nil, err
	}

	return &core.RedisCommand{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respond(cmd *core.RedisCommand, conn io.ReadWriter) {
	err := core.EvalAndRespond(cmd, conn)
	if err != nil {
		respondError(err, conn)
	}
}

func respondError(err error, conn io.ReadWriter) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}
