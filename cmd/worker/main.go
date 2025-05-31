package main

import (
	"flag"
	"log/slog"
	"os"
	"time"
)

func main() {
	queue := flag.String("queue", "amqp://guest:guest@localhost:5672/", "AMQP URL")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("starting worker", "queue", *queue)
	for range time.Tick(time.Second) {
		logger.Info("worker tick")
	}
}
