package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"traceable-ai-quiz/internal/application/health"
	"traceable-ai-quiz/internal/platform/config"
	"traceable-ai-quiz/internal/platform/logging"
	"traceable-ai-quiz/internal/transport/httpapi"
)

func main() { os.Exit(run()) }

func run() int {
	probe := flag.Bool("healthcheck", false, "check process liveness")
	flag.Parse()
	cfg, err := config.Load(os.Getenv, false)
	log := logging.New(os.Stdout, "api", cfg.Env, cfg.Commit, cfg.LogLevel)
	if err != nil {
		log.Error("startup_failed", "error_code", "CONFIG_INVALID", "message", err.Error())
		return 1
	}
	if *probe {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.APIPort)) + "/health")
		if err != nil {
			return 1
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return 1
		}
		return 0
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: net.JoinHostPort(cfg.ListenHost, strconv.Itoa(cfg.APIPort)), Handler: httpapi.New(health.New(cfg.Name, cfg.Version), log), ReadHeaderTimeout: cfg.ReadTimeout, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout}
	// Route the standard library's diagnostics through the same JSON writer.
	server.ErrorLog = slogAdapter(log)
	ended := make(chan error, 1)
	go func() { ended <- server.ListenAndServe() }()
	log.Info("server_starting", "port", cfg.APIPort)
	select {
	case err := <-ended:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server_failed", "error_code", "LISTEN_FAILED", "message", err.Error())
			return 1
		}
	case <-ctx.Done():
		log.Info("shutdown_started")
		closing, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(closing); err != nil {
			_ = server.Close()
			log.Error("shutdown_failed", "error_code", "SHUTDOWN_TIMEOUT")
			return 1
		}
		log.Info("shutdown_complete")
	}
	return 0
}
