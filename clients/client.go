package main

import (
	"fmt"
	"net"
	"log"
	"flag"
)

func main() {
	udp := flag.Bool("udp", false, "Use UDP instead of TCP")
	serverAddr := flag.String("server", "localhost:8080", "Server address")
	flag.Parse()

	var protocol string = ""
	if *udp {
		protocol = "udp"
	} else {
		protocol = "tcp"
	}

	conn, err := net.Dial(protocol, *serverAddr)
	if err != nil {
		log.Fatal("Failed to connect to load balancer");
	}
	defer conn.Close()

	message := "First message to load balancer"
	fmt.Fprintf(conn, message + "\n")

	buffer := make([]byte, 1024)

	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Failed to read from load balancer");
		}

		fmt.Println("Response from load balancer: ", string(buffer[:bytesRead]))
		fmt.Scanln(&message)
		fmt.Fprintf(conn, message + "\n")

	}
}