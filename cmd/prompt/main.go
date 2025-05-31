package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("listen", ":8090", "listen address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("starting prompt service", "addr", *addr)
	if err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			logger.Info("ping", "request_id", reqID)
		}
		w.Write([]byte("ok"))
	})); err != nil {
		logger.Error("server error", "err", err)
	}
}
