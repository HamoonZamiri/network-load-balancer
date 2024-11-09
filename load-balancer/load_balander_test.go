package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"example/network-load-balancer/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadBalancerWithTCP(t *testing.T) {
	// Start backend servers
	server1 := startTestServer(t, "localhost:8081")
	defer server1.Close()
	server2 := startTestServer(t, "localhost:8082")
	defer server2.Close()

	// Start load balancer
	config.Config = &config.AppConfig{}
	config.Config.Servers = []string{"localhost:8081", "localhost:8082"}
	config.Config.Protocol = "tcp"
	config.Config.BindAddress = "localhost:8080"
	go func() {
		if err := Run(); err != nil {
			t.Errorf("Load balancer error: %v", err)
		}
	}()
	time.Sleep(time.Second) // Give the load balancer time to start

	// Run test clients
	client1Response := runTestClient(t, "localhost:8080")
	client2Response := runTestClient(t, "localhost:8080")

	// Assert responses
	assert.NotEqual(t, client1Response, client2Response)
}

func startTestServer(t *testing.T, address string) net.Listener {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleTestConnection(conn, address)
		}
	}()

	return listener
}

func handleTestConnection(conn net.Conn, address string) {
	defer conn.Close()
	message := fmt.Sprintf("Response from %s", address)
	conn.Write([]byte(message))
}

func runTestClient(t *testing.T, address string) string {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect to load balancer: %v", err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from load balancer: %v", err)
	}

	return string(buffer[:n])
}
