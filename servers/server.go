package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
)

func handleTCPRequest(conn net.Conn) {
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

func handleUDPRequest(conn net.PacketConn, wg *sync.WaitGroup) {
	defer wg.Done()

	// Send a response to the client
	for {
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Printf("Failed to read from client: %s", err)
			return
		}
		request := string(buffer[:bytesRead])
		log.Printf("Received request from UDP client %s: %s", addr, request)
		conn.WriteTo([]byte(fmt.Sprintf("Hello from server, you sent '%s'\n", request)), addr)
	}
}

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:8080", "Server address")
	udp := flag.Bool("udp", false, "Use UDP instead of TCP")
	flag.Parse()

	var protocol string = ""
	if *udp {
		protocol = "udp"
	} else {
		protocol = "tcp"
	}

	var wg sync.WaitGroup

	// Create a listener to accept incoming connections
	if protocol == "tcp" {
		listener, err := net.Listen(protocol, *serverAddr)
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
			go handleTCPRequest(conn)
		}
	} else {
		conn, err := net.ListenPacket(protocol, *serverAddr)
		if err != nil {
			log.Fatalf("Failed to bind UDP: %s", err)
		}

		log.Printf("Server is listening on %s", *serverAddr)
		for {
			// Handle the client connection
			wg.Add(1)
			go handleUDPRequest(conn, &wg)

			// Wait for all goroutines to finish before exiting
			wg.Wait()
		}
	}
}
