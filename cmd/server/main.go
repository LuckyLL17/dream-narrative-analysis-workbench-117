package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dream117/config"
	"dream117/internal/httpapi"
	"dream117/internal/jobs"
	"dream117/internal/security"
	"dream117/internal/service"
	"dream117/internal/store"
	"dream117/internal/telemetry"
	"dream117/web"
)

func main() {
	cfg := config.Load()
	data, err := store.Open(
		cfg.DataFile)
	if err != nil {
		log.Fatal(err)
	}
	secret := envOr("DREAM_SECRET", randomSecret())
	codec := security.NewTokenCodec(secret, cfg.SessionTTL)
	auth := service.NewAuthService(data, codec)
	session := service.NewSessionService(auth)
	elements := service.NewElementService(data)
	dreams := service.NewDreamService(data, elements)
	analysisService := service.NewAnalysisService(data)
	reports := service.NewReportService(data)
	exports := service.NewExportService(data)
	health := service.NewHealthService(data)
	app := httpapi.NewApp(auth, session, dreams, elements, analysisService, reports, exports, health)
	metrics := &telemetry.Metrics{}
	handler := metrics.Observe(app.Handler(web.Handler()))
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 12 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	queue := jobs.NewQueue(data)
	refresher := jobs.NewRefresher(queue, reports)
	scheduler := jobs.NewScheduler(cfg.RefreshEvery, refresher)
	ctx, stop :=
		signal.NotifyContext(
			context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	scheduler.Start(ctx)
	log.Printf("dream workbench listening addr=%s data=%s", cfg.HTTPAddr, cfg.DataFile)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel :=
		context.WithTimeout(
			context.Background(), cfg.ShutdownWindow)
	defer cancel()
	_ = server.Shutdown(shutdown)
}

func envOr(
	key,
	fallback string,
) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func randomSecret() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "local-dream-secret"
	}
	return hex.EncodeToString(buf)
}
