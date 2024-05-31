package models

type Request struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}
