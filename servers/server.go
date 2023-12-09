package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func handleRequest(conn net.Conn) {
	defer conn.Close()

	// Send a response to the client
	response := "Hello from the server!"
	fmt.Fprintf(conn, response)
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
