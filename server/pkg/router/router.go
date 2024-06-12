package router

import (
	"encoding/json"
	"fmt"
	"log"

	. "udp-hole-punch/pkg/models"
)

type Router struct {
	methods map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		methods: make(map[string]HandlerFunc),
	}
}

func (r *Router) AddRoute(event string, handler HandlerFunc) *Router {
	log.Printf("Adding route for [%s]", event)
	r.methods[event] = handler
	return r
}

func (r *Router) HandleEvent(client *Client, bytesRead []byte) {
	log.Print("Handling event from ", client.GetRemoteAddr())
	var request Request

	err := json.Unmarshal(bytesRead, &request)
	if err != nil {
		_, err = client.GetConn().WriteToUDP([]byte("Invalid request"), client.GetRemoteAddr())
		if err != nil {
			log.Printf("Error writing to UDP: %v", err)
		}
		return
	}

	log.Printf("Handling event [%s]", request.Event)
	err = r.methods[request.Event](client, request.Payload)
	if err != nil {
		fmt.Println(err)
	}
}
