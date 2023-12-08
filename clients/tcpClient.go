package main

import (
	"fmt"
	"net"
	"log"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Failed to connect to load balancer");
	}
	defer conn.Close()

	message := "First message to load balancer"
	fmt.Fprintf(conn, message + "\n")

	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Fatal("Failed to read from load balancer");
	}
	fmt.Println("Response from load balancer: ", string(buffer[:bytesRead]))
}