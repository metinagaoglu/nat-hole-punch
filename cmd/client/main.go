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
	"time"

	"github.com/metnagaoglu/holepunch/peer"
)

func main() {
	signalAddr := flag.String("signal-address", "127.0.0.1:3986", "Signal server address")
	localAddr := flag.String("local-address", "0.0.0.0:0", "Local bind address (use port 0 for auto)")
	roomKey := flag.String("room-key", "default", "Room key for peer discovery")
	flag.Parse()

	client, err := peer.Connect(*signalAddr, *localAddr, *roomKey)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Listen()
	}()

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

func interactiveLoop(client *peer.Client) {
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
			peers := client.Peers().GetAll()
			if len(peers) == 0 {
				fmt.Println("No peers connected")
			} else {
				fmt.Printf("Connected peers (%d):\n", len(peers))
				for _, p := range peers {
					fmt.Printf("  - %s (last seen: %s ago)\n", p.Addr, time.Since(p.LastSeen).Round(time.Second))
				}
			}

		case strings.HasPrefix(line, "/send "):
			filePath := strings.TrimSpace(strings.TrimPrefix(line, "/send "))
			if filePath == "" {
				fmt.Println("Usage: /send <filepath>")
				continue
			}
			client.SendFile(filePath)

		case line == "/quit":
			client.Shutdown()
			os.Exit(0)

		default:
			client.Send(line)
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
