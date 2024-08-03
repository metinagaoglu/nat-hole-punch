package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type RegisterRequest struct {
	LocalIp string `json:"local_ip"`
	Key     string `json:"key"`
}

type Event struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

func main() {

	signalAddressPtr := flag.String("signal-address", "127.0.0.1:3986", "Signal server address")
	localAddressPtr := flag.String("local-address", "127.0.0.1:4000", "Local address")
	roomKeyAddressPtr := flag.String("room-key", "default", "Room key")

	flag.Parse()

	signalAddress := *signalAddressPtr
	localAddress := *localAddressPtr
	roomKey := *roomKeyAddressPtr

	remote, err := net.ResolveUDPAddr("udp", signalAddress)
	if err != nil {
		panic(err)
	}
	local, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", local)
	if err != nil {
		panic(err)
	}

	register, _ := json.Marshal(RegisterRequest{
		LocalIp: local.String(),
		Key:     roomKey,
	})

	go func() {
		jsonRegister, _ := json.Marshal(Event{
			Event:   "register",
			Payload: string(register),
		})

		log.Println("registering ", string(jsonRegister))
		bytesWritten, err := conn.WriteTo([]byte(jsonRegister), remote)
		if err != nil {
			panic(err)
		}

		log.Println(bytesWritten, " bytes written")
	}()

	listen(conn, local.String())
}

func listen(conn *net.UDPConn, local string) {
	for {
		log.Println("listening")
		buffer := make([]byte, 1024)
		bytesRead, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("[ERROR]", err)
			continue
		}

		log.Println("[INCOMING]", string(buffer[0:bytesRead]))
		if string(buffer[0:bytesRead]) == "Hello!" {
			log.Println("received Hello!", " from ", addr.String(), " sending Hi! 🚀🚀🚀 to ", addr.String())
			continue
		}

		for _, a := range strings.Split(string(buffer[0:bytesRead]), ",") {
			if a != local {
				go chatter(conn, a)
			}
		}
	}
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)

	// Get Input from user
	/*
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
	*/

	for {
		conn.WriteTo([]byte("Hello!"), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
