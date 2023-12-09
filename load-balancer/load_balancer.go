package main

import (
	"flag"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// Represents a pool of backend servers for load balancing
type BackendPool struct {
	servers []string
	counter int
	mu      sync.Mutex
}

// Selects a backend server from the pool using a round-robin strategy
func (pool *BackendPool) Choose() string {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	idx := pool.counter % len(pool.servers)
	pool.counter++
	return pool.servers[idx]
}

// Returns a string representation of the backend servers
func (pool *BackendPool) String() string {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return strings.Join(pool.servers, ", ")
}

var (
	bind    = flag.String("bind", "", "The address to bind on")
	balance = flag.String("balance", "", "The backend servers to balance connections across, separated by commas")
	pool    *BackendPool
)

func init() {
	flag.Parse()

	// Checks if bind flag is empty or is not provided
	if *bind == "" {
		log.Fatalln("Specify address to listen on with -bind")
	}

	// Checks if balance flag is empty
	servers := strings.Split(*balance, ",")
	if len(servers) == 1 && servers[0] == "" {
		log.Fatalln("Specify backend servers with -balance if you want to use load balancing")
	}

	pool = &BackendPool{servers: servers}
}

// Handles a new incoming connection by selecting a backend server and establishing a connection
func handleConnection(clientConn net.Conn, backendPool *BackendPool) {
	serverAddr := backendPool.Choose()
	log.Printf("Routing connection to server: %s", serverAddr)

	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		clientConn.Close()
		log.Printf("Failed to dial %s: %s", serverAddr, err)
		return
	}

	// Copy data from client to backend server
	if _, err := io.Copy(serverConn, clientConn); err != nil {
		log.Printf("Error copying from client to server: %s", err)
	}

	// Copy data from backend server to client
	if _, err := io.Copy(clientConn, serverConn); err != nil {
		log.Printf("Error copying from server to client: %s", err)
	}

	// Close connections after copying is complete or in case of error
	clientConn.Close()
	serverConn.Close()

	log.Printf("Connection completed for server: %s", serverAddr)
}

// Entry point of load balancer program
func main() {
	// Create a listener to accept incoming TCP connections
	listener, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatalf("Failed to bind: %s", err)
	}

	// Log information about the program's state
	log.Printf("Listening on %s, balancing connections across: %s", *bind, pool)

	// Continuously accept incoming client connections
	for {
		// Accept a new client connection
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept: %s", err)
			continue
		}

		// Handle the client connection concurrently by selecting a backend server and establishing a connection
		go handleConnection(clientConn, pool)
	}
}
