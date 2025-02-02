package main

import (
	"github.com/maximkir777/word_of_wisdom/internal/client"
	"github.com/maximkir777/word_of_wisdom/internal/config"
)

func main() {
	cfg := config.NewClientConfig()

	// Initialize and run the client
	cl := client.NewClient(cfg.ServerAddr, cfg.Timeout)
	cl.Run()
}
