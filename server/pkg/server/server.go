package server

import (
	"log"
	"net"

	. "udp-hole-punch/pkg/models"
	. "udp-hole-punch/pkg/router"
)

type UDPServer struct {
	conn    *net.UDPConn
	clients map[string]*Client
	router  *Router
}

func NewUDPServer() *UDPServer {
	return &UDPServer{
		conn:    nil,
		clients: map[string]*Client{},
	}
}

func (u *UDPServer) SetRoutes(r *Router) *UDPServer {
	u.router = r
	return u
}

func (u *UDPServer) Bind(port int) *UDPServer {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}

	u.conn = conn
	return u
}

func (u *UDPServer) Listen() {
	log.Printf("Listening on %s 🚀🚀🚀", u.conn.LocalAddr().String())
	for {
		buffer := make([]byte, 1024)
		bytesRead, remoteAddr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalf("Error reading from UDP: %v", err)
			panic(err)
		}

		log.Printf("Received %s from %s", string(buffer[0:bytesRead]), remoteAddr)
		client := NewClient().SetRemoteAddr(remoteAddr).SetCreateAt().SetConn(u.conn)
		u.clients[remoteAddr.String()] = client

		u.router.HandleEvent(client, buffer[0:bytesRead])
	}
}
