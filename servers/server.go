package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func handleRequest(conn net.Conn) {
	// Send a response to the client
	for {
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Failed to read from client: %s", err)
			return
		}
		request := string(buffer[:bytesRead])
		log.Printf("Received request from client: %s", request)
		fmt.Fprintf(conn, "Hello from server, you sent '%s'\n", request)
	}
}

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:8080", "Server address")
	flag.Parse()

	// Create a listener to accept incoming TCP connections
	listener, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("Failed to bind: %s", err)
	}

	log.Printf("Server is listening on %s", *serverAddr)

	for {
		// Accept a new client connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept: %s", err)
			continue
		}

		// Handle the client connection
		go handleRequest(conn)
	}
}
