package main

import (
	"log"
	"net"
	"syscall"

	"github.com/NitinReddy01/redis/config"
	"github.com/NitinReddy01/redis/core"
)

var conn_clients = 0

func RunAsyncTCPServer() error {
	max_clients := 20000

	var events []syscall.Kevent_t = make([]syscall.Kevent_t, max_clients)

	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	ipv4 := net.ParseIP(config.HOST).To4()
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.PORT,
		Addr: [4]byte(ipv4),
	}); err != nil {
		return err
	}

	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	kqFD, err := syscall.Kqueue()
	if err != nil {
		return err
	}
	defer syscall.Close(kqFD)

	var serverSocketEvent syscall.Kevent_t = syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD,
	}

	_, err = syscall.Kevent(kqFD, []syscall.Kevent_t{serverSocketEvent}, nil, nil)
	if err != nil {
		return err
	}

	for {
		nevents, err := syscall.Kevent(kqFD, nil, events, nil)
		if err != nil {
			continue
		}

		for i := range nevents {
			if events[i].Ident == uint64(serverFD) {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err occured", err)
					continue
				}

				conn_clients++
				syscall.SetNonblock(fd, true)
				log.Printf("client connected, total: %d", conn_clients)

				var socketClientEvent syscall.Kevent_t = syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
				}

				_, err = syscall.Kevent(kqFD, []syscall.Kevent_t{socketClientEvent}, nil, nil)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDComm{Fd: int(events[i].Ident)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Ident))
					conn_clients--
					log.Printf("client disconnected, total: %d", conn_clients)
					continue
				}
				respond(cmd, comm)
			}
		}
	}
}
