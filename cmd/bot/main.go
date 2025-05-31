package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("listen", ":8080", "listen address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	secret := os.Getenv("TELEGRAM_SECRET_TOKEN")
	if secret == "" {
		logger.Warn("TELEGRAM_SECRET_TOKEN not set")
	}
	logger.Info("starting bot", "addr", *addr)
	if err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkSecretToken(r, secret, logger) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			logger.Info("ping", "request_id", reqID)
		}
		w.Write([]byte("ok"))
	})); err != nil {
		logger.Error("server error", "err", err)
	}
}
