package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jb843051627/avalanche-net/internal/api"
	"github.com/jb843051627/avalanche-net/internal/service"
	"github.com/jb843051627/avalanche-net/internal/store"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	dbPath := envOr("AVALANCHE_NET_DB", "data/avalanche-net.db")
	addr := envOr("AVALANCHE_NET_ADDR", ":8080")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st,
		service.WithHeartbeatConfig(5, 30),
		service.WithCacheKeep(512),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stopWatcher := svc.StartHeartbeatWatcher(ctx)
	defer stopWatcher()

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.New(svc),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("avalanche-net listening on %s (db=%s)", addr, dbPath)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
	log.Printf("avalanche-net shut down cleanly")
}
