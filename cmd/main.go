package main

import "log"

func main() {
	// RunSyncTCPServer()
	if err := RunAsyncTCPServer(); err != nil {
		log.Fatal(err)
	}
}
