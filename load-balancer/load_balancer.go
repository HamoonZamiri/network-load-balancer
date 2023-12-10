package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Represents a pool of backend servers for load balancing
type BackendPool struct {
	servers []string
	counter int
	mu      sync.Mutex
	healthChecks map[string]bool
}

func (pool *BackendPool) HealthCheck() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, server := range pool.servers {
		conn, err := net.Dial("tcp", server)
		prev := pool.healthChecks[server]

		if err != nil {
			pool.healthChecks[server] = false
			log.Printf("Server %s is down", server)
		} else {
			pool.healthChecks[server] = true
			fmt.Fprintf(conn, "Health Check\n")
		}

		if prev && !pool.healthChecks[server] {
			log.Printf("Server %s has gone down", server)
		}
		if !prev && pool.healthChecks[server] {
			log.Printf("Server %s has come back up", server)
		}
		if conn != nil {
			conn.Close()
		}
	}

}

func (pool *BackendPool) HealthCheckLoop() {
	const healthCheckInterval = 5 * time.Second
	for {
		pool.HealthCheck()
		time.Sleep(healthCheckInterval)
	}
}

// Selects a backend server from the pool using a round-robin strategy
func (pool *BackendPool) Choose() string {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	idx := pool.counter % len(pool.servers)
	pool.counter++
	for !pool.healthChecks[pool.servers[idx]] {
		idx = pool.counter % len(pool.servers)
		pool.counter++
	}

	return pool.servers[idx]
}

// Returns a string representation of the backend servers
func (pool *BackendPool) String() string {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return strings.Join(pool.servers, ", ")
}


func (pool *BackendPool) initializeHealthCheck() {
	pool.healthChecks = make(map[string]bool)
	for _, server := range pool.servers {
		pool.healthChecks[server] = false
	}
}

// Handle TCP connection
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

// Keep track of client-to-server mapping separately for UDP
type UDPClientToServerMap struct {
	mu             sync.Mutex
	clientToServer map[string]string
}

// Choose a backend server based on the client's address
func chooseServer(pool *BackendPool, clientAddr string, udpClientMap *UDPClientToServerMap) string {
	// Check if the client has been assigned a server for UDP
	udpClientMap.mu.Lock()
	serverAddr, ok := udpClientMap.clientToServer[clientAddr]
	udpClientMap.mu.Unlock()

	if ok {
		return serverAddr
	}

	// Choose a backend server using round-robin
	return pool.Choose()
}

// Handle UDP connection
func handleUDPConnection(clientConn net.PacketConn, backendPool *BackendPool, udpClientMap *UDPClientToServerMap, wg *sync.WaitGroup) {
	defer wg.Done()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := clientConn.ReadFrom(buffer)
		if err != nil {
			log.Printf("Error reading from UDP client %s: %s", addr, err)
			break
		}

		// Choose a backend server based on the client's address
		serverAddr := chooseServer(backendPool, addr.String(), udpClientMap)

		// Map client address to chosen server address for UDP
		udpClientMap.mu.Lock()
		udpClientMap.clientToServer[addr.String()] = serverAddr
		udpClientMap.mu.Unlock()

		// Dial a connection to the selected backend server
		serverConn, err := net.Dial("udp", serverAddr)
		if err != nil {
			log.Printf("Failed to dial %s: %s", serverAddr, err)
			continue
		}

		// Forward the client's message to the server
		_, err = serverConn.Write(buffer[:n])
		if err != nil {
			log.Printf("Error writing to UDP server %s: %s", serverAddr, err)
		}

		// Reset the buffer before reading the response
		buffer = make([]byte, 1024)

		// Read the response from the server
		_, err = serverConn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from UDP server %s: %s", serverAddr, err)
		}

		// Send the server's response back to the client
		_, err = clientConn.WriteTo(buffer, addr)
		if err != nil {
			log.Printf("Error writing to UDP client %s: %s", addr, err)
		}

		// Close the server connection
		serverConn.Close()

		log.Printf("Forwarded message from %s to server %s and back to %s", addr, serverAddr, addr)
	}
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
	pool.initializeHealthCheck()

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
      // Every 5 seconds execute the health check in a go routine concurrently
      go pool.HealthCheckLoop()

      clientConn, err := listener.Accept()
      if err != nil {
        log.Printf("Failed to accept: %s", err)
        continue
      }

      // Handle the client connection concurrently by selecting a backend server and establishing a connection
      go handleTCPConnection(clientConn, pool)
      }
	} else {
		  // Create UDPClientToServerMap
		  udpClientMap := &UDPClientToServerMap{
			  clientToServer: make(map[string]string),
		  }
		
      clientConn, err := net.ListenPacket(protocol, *bind)

      if err != nil {
        log.Fatalf("Failed to bind UDP: %s", err)
      }

      // Log information about the program's state
      log.Printf("Listening on %s, balancing connections across: %s", *bind, pool)

      for {
        // Handle the client connection
        wg.Add(1)
        go handleUDPConnection(clientConn, pool, udpClientMap, &wg)

        // Wait for all goroutines to finish before exiting
        wg.Wait()
      }
	}
}
