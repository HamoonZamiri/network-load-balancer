package main

import (
	"fmt"
	"net"
	"log"
	"flag"
)

func main() {
	flag.Parse()
	udp := flag.Bool("udp", false, "Use UDP instead of TCP")
	var protocol string = ""
	if *udp {
		protocol = "udp"
	} else {
		protocol = "tcp"
	}

	conn, err := net.Dial(protocol, "localhost:8080")
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
			log.Fatal("Failed to read from load balancer");
		}

		fmt.Println("Response from load balancer: ", string(buffer[:bytesRead]))
		fmt.Scanln(&message)
		fmt.Fprintf(conn, message + "\n")

	}
}