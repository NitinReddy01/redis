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

	// buffer where kqueue will put events for FDs that are ready for IO
	// we allocate for max_clients but kqueue only fills the ones that are ready
	// and returns the count in nevents
	var events []syscall.Kevent_t = make([]syscall.Kevent_t, max_clients)

	// creating a raw TCP socket instead of net.Listen because
	// we need the raw file descriptor to work with kqueue directly
	// AF_INET = IPv4, SOCK_STREAM = TCP (persistent connection)
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	// non-blocking so accept() and read() return immediately
	// instead of blocking the thread when there's nothing to do
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// bind the socket to an IP and PORT
	// without this the socket exists but nobody can connect to it
	ipv4 := net.ParseIP(config.HOST).To4()
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.PORT,
		Addr: [4]byte(ipv4),
	}); err != nil {
		return err
	}

	// tell the OS this socket is ready to accept incoming connections
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// create a kqueue instance (equivalent of epoll_create1 on Linux)
	// kqueue lets us monitor multiple FDs without blocking on any single one
	// instead of us watching sockets, the OS watches them and tells us
	// when something is ready — this is IO multiplexing
	kqFD, err := syscall.Kqueue()
	if err != nil {
		return err
	}
	defer syscall.Close(kqFD)

	// register the server socket with kqueue
	// EVFILT_READ = notify when there's data to read (equivalent of EPOLLIN)
	// for a server socket, "data to read" means a new client wants to connect
	// EV_ADD = add this FD to kqueue's watch list
	var serverSocketEvent syscall.Kevent_t = syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD,
	}
	_, err = syscall.Kevent(kqFD, []syscall.Kevent_t{serverSocketEvent}, nil, nil)
	if err != nil {
		return err
	}

	// the event loop — this is the heart of IO multiplexing
	// single thread, no goroutines, no mutexes
	// kqueue watches all registered FDs and wakes us up when any are ready
	for {
		// blocks efficiently until at least one FD has IO ready
		// the OS does the watching via hardware interrupts, not polling
		// when data arrives on any socket, OS wakes this call up
		nevents, err := syscall.Kevent(kqFD, nil, events, nil)
		if err != nil {
			// can fail for non-critical reasons like signal interrupts
			continue
		}

		for i := range nevents {
			if events[i].Ident == uint64(serverFD) {
				// server socket is ready = a new client wants to connect
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err occured", err)
					continue
				}

				conn_clients++
				syscall.SetNonblock(fd, true)
				log.Printf("client connected, total: %d", conn_clients)

				// register this client's socket with kqueue too
				// now kqueue watches both the server (for new connections)
				// and this client (for incoming commands)
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
				// client socket is ready = client has sent data
				// read the RESP command, execute it, send response
				// all still on the same single thread
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
