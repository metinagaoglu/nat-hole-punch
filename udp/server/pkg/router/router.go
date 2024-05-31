package router

import (
	"encoding/json"
	"fmt"
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
	r.methods[event] = handler
	return r
}

func (r *Router) HandleEvent(client *Client, bytesRead []byte) {
	fmt.Println()
	fmt.Println("Router is handling event")
	var request Request

	json.Unmarshal(bytesRead, &request)
	fmt.Println(request.Event)
	fmt.Println(string(request.Payload))
	r.methods[request.Event](client, request.Payload)
}
