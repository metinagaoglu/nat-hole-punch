package server

import (
	"fmt"
	"log"
	"net"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/handlers"
	. "udp-hole-punch/pkg/models"
	. "udp-hole-punch/pkg/router"
)

type UDPServer struct {
	conn       *net.UDPConn
	clients    map[string]*Client
	router     *Router
	config     *config.Config
	bufferSize int
}

func NewUDPServer(cfg *config.Config) *UDPServer {
	return &UDPServer{
		conn:       nil,
		clients:    map[string]*Client{},
		config:     cfg,
		bufferSize: cfg.BufferSize,
	}
}

func (u *UDPServer) SetRoutes(r *Router) *UDPServer {
	u.router = r
	return u
}

func (u *UDPServer) Bind() (*UDPServer, error) {
	addr := net.UDPAddr{
		Port: u.config.ServerPort,
		IP:   net.ParseIP(u.config.ServerHost),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}

	u.conn = conn
	return u, nil
}

func (u *UDPServer) Listen() error {
	if u.conn == nil {
		return fmt.Errorf("server not bound to any address")
	}

	// Set server connection for handlers to use
	// This allows handlers to send messages using server's connection
	handlers.SetServerConnection(u.conn)

	log.Printf("Listening on %s 🚀🚀🚀", u.conn.LocalAddr().String())
	for {
		buffer := make([]byte, u.bufferSize)
		bytesRead, remoteAddr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		log.Printf("Received %s from %s", string(buffer[0:bytesRead]), remoteAddr)
		client := NewClient().SetRemoteAddr(remoteAddr).SetCreateAt().SetConn(u.conn)
		u.clients[remoteAddr.String()] = client

		u.router.HandleEvent(client, buffer[0:bytesRead])
	}
}

