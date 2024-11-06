package config

import (
	"flag"
	"log"
	"os"
	"strings"
)

type (
	AppConfig struct {
		Protocol    string
		BindAddress string
		Servers     []string
	}
)

var Config *AppConfig = nil

func getOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func SetConfig() {
	if Config != nil {
		return
	}

	// Parse command line flags
	bind := flag.String("bind", "", "The address to bind on")
	balance := flag.String("balance", "", "The backend servers to balance connections across, separated by commas")
	udp := flag.Bool("udp", false, "Use UDP instead of TCP")
	flag.Parse()

	// Check if bind flag is empty or is not provided
	if *bind == "" {
		log.Fatalln("Specify address to listen on with -bind")
	}

	// Check if balance flag is empty
	servers := strings.Split(*balance, ",")
	if len(servers) == 1 && servers[0] == "" || len(servers) == 0 {
		log.Fatalln("Specify backend servers with -balance")
	}

	Config = &AppConfig{
		Protocol:    "tcp",
		BindAddress: *bind,
		Servers:     servers,
	}
	if *udp {
		Config.Protocol = "udp"
	}
}
