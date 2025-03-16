package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"strconv"
	"sync"
	"time"

	"github.com/CGuilhem/hackathon-notifier/backend/internal/server"
	"github.com/CGuilhem/hackathon-notifier/backend/internal/websocket"
	"github.com/CGuilhem/hackathon-notifier/backend/logger"
	"github.com/joho/godotenv"
)

func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Logger initialization
	logger := logger.NewLogger(stdout)
	logger.Info("Hackathon-notifier server start-up")

	// Websockets manager initialization
	wsManager := websocket.NewManager(logger)

	// Http server initialization
	srv := server.NewServer(
		wsManager,
		getenv,
		logger,
	)
	apiTimeout, err := strconv.Atoi(getenv("API_TIMEOUT"))
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr: "0.0.0.0:3000",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(apiTimeout),
		ReadTimeout:  time.Second * time.Duration(apiTimeout),
		IdleTimeout:  time.Second * time.Duration(apiTimeout),
		Handler:      srv,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("Error listening and serving: %s", err))
		}
		logger.Info(fmt.Sprintf("Listening on %s", httpServer.Addr))
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down http server: %s\n", err))
		}
		logger.Info("Gracefully shutting down")
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()

	if os.Getenv("ENV") == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		if err := godotenv.Load(dir + "/.env"); err != nil {
			log.Fatal(err)
		}
	}

	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
