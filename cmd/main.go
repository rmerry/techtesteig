package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	btcore "github.com/rmerry/btcorehandshaker/internal/btcore"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	client := btcore.NewClient("127.0.0.1", 18444)
	logger.Info("starting bitcoin core client")
	go func() {
		if err := client.Connect(ctx); err != nil {
			if !errors.Is(err, btcore.ErrContext) {
				logger.Error("connection error", "err", err)
				os.Exit(1)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("interrupt signal received")
	logger.Info("shutting down client")
	client.Disconnect()
}
