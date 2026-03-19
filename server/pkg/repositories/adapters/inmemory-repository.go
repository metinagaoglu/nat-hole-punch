package adapters

// import redis
import (
	"fmt"

	. "udp-hole-punch/pkg/models"
)

type InMemoryRepository struct {
	clients map[string][]*Client
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		clients: map[string][]*Client{},
	}
}

func (r *InMemoryRepository) AddClient(key string, client *Client, ttl int32) error {
	r.clients[key] = append(r.clients[key], client)
	return nil
}

func (r *InMemoryRepository) RemoveClient(key string, client *Client) error {
	// remove client from clients by key
	fmt.Println(r.clients)
	for i, c := range r.clients[key] {
		fmt.Println(c.GetRemoteAddr())
		fmt.Println(client.GetRemoteAddr())
		fmt.Println(client.CompareAddr(c.GetRemoteAddr()))
		fmt.Println()
		if client.CompareAddr(c.GetRemoteAddr()) {

			fmt.Println("Removing client from clients by key")
			r.clients[key] = append(r.clients[key][:i], r.clients[key][i+1:]...)
			break
		}
	}
	fmt.Println(r.clients)
	return nil
}

func (r *InMemoryRepository) GetClients() ([]*Client, error) {
	return nil, nil
}

func (r *InMemoryRepository) GetClientsByKey(key string) ([]*Client, error) {
	return r.clients[key], nil
}
