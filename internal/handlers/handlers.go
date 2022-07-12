package handlers

import (
	"context"

	"shortener/internal/storage"
)

type Storage interface {
	Push(string, string) (string, error)
	Get(string) (string, bool)
	Users(string, string) ([]storage.Users, error)
	Ping(context.Context) error
	Batch(string, []*Batch) error
	Close() error
}

type Message struct {
	URL    string `json:"url,omitempty"`
	Result string `json:"result,omitempty"`
}

type Batch struct {
	ID    string `json:"correlation_id"`
	URI   string `json:"original_url,omitempty"`
	Short string `json:"short_url,omitempty"`
}
