package main

import (
	. "udp-hole-punch/pkg/router"
	. "udp-hole-punch/pkg/server"
)

func main() {
	NewUDPServer().SetRoutes(InıtializeRoutes()).Bind(3986).Listen()
}
