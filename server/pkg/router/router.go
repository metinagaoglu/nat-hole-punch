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

	err := json.Unmarshal(bytesRead, &request)
	if err != nil {
		_, err = client.GetConn().WriteToUDP([]byte("Invalid request"), client.GetRemoteAddr())
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	fmt.Println(request.Event)
	fmt.Println(string(request.Payload))
	err = r.methods[request.Event](client, request.Payload)
	if err != nil {
		fmt.Println(err)
	}
}
