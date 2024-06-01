package server

import (
	"fmt"
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
	fmt.Println("Server is listening on port 3986")
	for {
		buffer := make([]byte, 1024)
		bytesRead, remoteAddr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		//		fmt.Println("Received ", string(buffer[0:bytesRead]), " from ", remoteAddr)
		client := NewClient().SetRemoteAddr(remoteAddr).SetCreateAt().SetConn(u.conn)
		u.clients[remoteAddr.String()] = client

		u.router.HandleEvent(client, buffer[0:bytesRead])
	}
}
