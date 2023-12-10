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

// Handles a new incoming connection by selecting a backend server and establishing a connection
func handleTCPConnection(clientConn net.Conn, backendPool *BackendPool) {
	serverAddr := backendPool.Choose()
	log.Printf("Routing connection to server: %s", serverAddr)

	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		clientConn.Close()
		log.Printf("Failed to dial %s: %s", serverAddr, err)
		return
	}

	// Use WaitGroup to wait for copying goroutines to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Copy data between client and backend server
	go func() {
		if _, err := io.Copy(serverConn, clientConn); err != nil {
			log.Printf("Error copying from client to server: %s", err)
		}
		wg.Done()
	}()

	go func() {
		if _, err := io.Copy(clientConn, serverConn); err != nil {
			log.Printf("Error copying from server to client: %s", err)
		}
		wg.Done()
	}()

	// Wait for copying goroutines to complete
	wg.Wait()

	// Close connections after copying is complete or in case of error
	clientConn.Close()
	serverConn.Close()

	log.Printf("Connection completed for server: %s", serverAddr)
}

// Handles a new incoming connection by selecting a backend server and establishing a connection
func handleUDPConnection(clientConn net.PacketConn, backendPool *BackendPool, wg *sync.WaitGroup) {
	defer wg.Done()

	serverAddr := backendPool.Choose()
	log.Printf("Routing connection to server: %s", serverAddr)

	serverConn, err := net.Dial("udp", serverAddr)
	if err != nil {
		clientConn.Close()
		log.Printf("Failed to dial %s: %s", serverAddr, err)
		return
	}

	// Use WaitGroup to wait for copying goroutines to complete
	var wgc sync.WaitGroup
	wgc.Add(2)

	// Copy data between client and backend server
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := clientConn.ReadFrom(buf)
			if err != nil {
				log.Printf("Error reading from UDP client %s: %s", addr, err)
				break
			}
			_, err = serverConn.Write(buf[:n])
			if err != nil {
				log.Printf("Error writing to UDP server %s: %s", addr, err)
				break
			}
		}
		wgc.Done()
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := serverConn.Read(buf)
			if err != nil {
				log.Printf("Error reading from UDP server: %s", err)
				break
			}
			_, err = clientConn.WriteTo(buf[:n], serverConn.LocalAddr())
			if err != nil {
				log.Printf("Error writing to UDP client: %s", err)
				break
			}
		}
		wgc.Done()
	}()

	// Wait for copying goroutines to complete
	wgc.Wait()

	// Close connections after copying is complete or in case of error
	clientConn.Close()
	serverConn.Close()

	log.Printf("Connection completed for server: %s", serverAddr)
}

// Entry point of load balancer program
func main() {
	// Parse command line flags
	bind := flag.String("bind", "", "The address to bind on")
	balance := flag.String("balance", "", "The backend servers to balance connections across, separated by commas")
	udp := flag.Bool("udp", false, "Use UDP instead of TCP")
	flag.Parse()

	var protocol string = ""
	if *udp {
		protocol = "udp"
	} else {
		protocol = "tcp"
	}

	// Check if bind flag is empty or is not provided
	if *bind == "" {
		log.Fatalln("Specify address to listen on with -bind")
	}

	// Check if balance flag is empty
	servers := strings.Split(*balance, ",")
	if len(servers) == 1 && servers[0] == "" {
		log.Fatalln("Specify backend servers with -balance")
	}

	// Create the backend pool
	pool := &BackendPool{servers: servers}

	var wg sync.WaitGroup

	// Create a listener to accept incoming connections
	if protocol == "tcp" {
		listener, err := net.Listen(protocol, *bind)
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
			go handleTCPConnection(clientConn, pool)
		}
	} else {
		clientConn, err := net.ListenPacket(protocol, *bind)
		if err != nil {
			log.Fatalf("Failed to bind UDP: %s", err)
		}

		// Log information about the program's state
		log.Printf("Listening on %s, balancing connections across: %s", *bind, pool)

		// Handle the client connection
		wg.Add(1)
		go handleUDPConnection(clientConn, pool, &wg)

		// Wait for all goroutines to finish before exiting
		wg.Wait()
	}
}
