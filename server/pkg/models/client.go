package models

import (
	"net"
	"time"
)

type Client struct {
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	createAt   int64
}

func NewClient() *Client {
	return &Client{
		remoteAddr: nil,
		createAt:   0,
	}
}

func (c *Client) SetConn(conn *net.UDPConn) *Client {
	c.conn = conn
	return c
}

func (c *Client) GetConn() *net.UDPConn {
	return c.conn
}

func (c *Client) SetRemoteAddr(addr *net.UDPAddr) *Client {
	c.remoteAddr = addr
	return c
}

func (c *Client) SetCreateAt() *Client {
	c.createAt = time.Now().Unix()
	return c
}

func (c *Client) GetRemoteAddr() *net.UDPAddr {
	return c.remoteAddr
}

func (c *Client) GetCreateAt() int64 {
	return c.createAt
}
