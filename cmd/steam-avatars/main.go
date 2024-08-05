package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mrmarble/steam-avatars/internal/database"
	"github.com/mrmarble/steam-avatars/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log := zerolog.New(output).With().Timestamp().Logger()

	db, err := database.OpenDB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	steamAPIKey := flag.String("key", "", "Steam API key")
	flag.Parse()

	if *steamAPIKey == "" {
		if key, ok := os.LookupEnv("STEAM_API_KEY"); ok {
			*steamAPIKey = key
		} else {
			log.Fatal().Msg("Steam API key is required")
		}
	}

	server := server.NewServer(log, db, *steamAPIKey)

	// Start server
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msg("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("shutting down the server")
	}
	if err := db.Close(); err != nil {
		log.Fatal().Err(err).Msg("closing the database")
	}
}
