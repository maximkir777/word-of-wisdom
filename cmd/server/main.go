package main

import (
	"github.com/maximkir777/word_of_wisdom/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/maximkir777/word_of_wisdom/internal/handler"
	"github.com/maximkir777/word_of_wisdom/internal/server"
	"github.com/maximkir777/word_of_wisdom/pkg/pow"
	"github.com/maximkir777/word_of_wisdom/pkg/wow"
)

func main() {
	cfg := config.NewServerConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize PoW with configuration values
	powInstance := pow.NewPoW(
		cfg.PowBaseDifficulty,
		cfg.PowMaxDifficulty,
		cfg.PowWindowSize,
		cfg.PowWindowDuration,
	)
	powInstance.StartDifficultyAdjuster()

	// Initialize wisdom service and request handler
	wisdomService := wow.NewService()
	handlerImpl := handler.NewWowHandler(wisdomService, powInstance)

	// Create and start the server using cfg.ListenAddr
	srv := server.NewServer(cfg.ListenAddr, handlerImpl, logger)
	if err := srv.Start(); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown on OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")
	srv.Stop()
}
