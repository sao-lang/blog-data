package es

import (
	"github.com/elastic/go-elasticsearch/v8"
)

type Client struct {
	ES *elasticsearch.Client
}

type Config struct {
	Addresses []string
	Username  string
	Password  string
}

func New(cfg Config) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	return &Client{ES: es}, nil
}
