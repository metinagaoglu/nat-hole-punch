package router

import (
	"encoding/json"
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
		// Only try to write back if client has a valid connection
		if client.GetConn() != nil && client.GetRemoteAddr() != nil {
			_, err = client.GetConn().WriteToUDP([]byte("Invalid request"), client.GetRemoteAddr())
			if err != nil {
				log.Printf("Error writing to UDP: %v", err)
			}
		}
		return
	}

	log.Printf("Handling event [%s]", request.Event)
	// check if the event is registered
	if _, ok := r.methods[request.Event]; !ok {
		log.Printf("Event not found [%s]", request.Event)
		return
	}
	err = r.methods[request.Event](client, request.Payload)
	if err != nil {
		log.Printf("Handler error for event [%s]: %v", request.Event, err)
	}
}
