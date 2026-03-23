package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	defaultSignalAddress = "127.0.0.1:3986"
	defaultLocalAddress  = "0.0.0.0:0" // OS assigns a free port
	defaultRoomKey       = "default"
)

func main() {
	signalAddr := flag.String("signal-address", defaultSignalAddress, "Signal server address")
	localAddr := flag.String("local-address", defaultLocalAddress, "Local bind address (use port 0 for auto)")
	roomKey := flag.String("room-key", defaultRoomKey, "Room key for peer discovery")
	flag.Parse()

	client, err := NewClient(*signalAddr, *localAddr, *roomKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Register(); err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start listener
	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Listen()
	}()

	// Start interactive input
	go interactiveLoop(client)

	printHelp()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		client.Shutdown()
	case err := <-errChan:
		if err != nil {
			log.Printf("Listen error: %v", err)
		}
	}
}

func interactiveLoop(client *Client) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case line == "/help":
			printHelp()

		case line == "/peers":
			client.peers.PrintPeers()

		case strings.HasPrefix(line, "/send "):
			filePath := strings.TrimPrefix(line, "/send ")
			filePath = strings.TrimSpace(filePath)
			if filePath == "" {
				fmt.Println("Usage: /send <filepath>")
				continue
			}
			client.SendFile(filePath)

		case line == "/quit":
			client.Shutdown()
			os.Exit(0)

		default:
			// Treat everything else as a text message
			client.SendText(line)
		}
	}
}

func printHelp() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        UDP Hole Punch Client             ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Println("║  <message>     Send text to all peers    ║")
	fmt.Println("║  /send <file>  Send file to all peers    ║")
	fmt.Println("║  /peers        List connected peers      ║")
	fmt.Println("║  /help         Show this help            ║")
	fmt.Println("║  /quit         Disconnect and exit       ║")
	fmt.Println("╚══════════════════════════════════════════╝")
}
