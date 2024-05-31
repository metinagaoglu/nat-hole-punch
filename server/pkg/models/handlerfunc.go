package models

type HandlerFunc func(client *Client, payload string) error
