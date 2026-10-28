package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	"traceable-ai-quiz/internal/platform/config"
	"traceable-ai-quiz/internal/platform/logging"
)

//go:embed nginx.conf.tmpl
var nginxTemplate string

func main() { os.Exit(run()) }
func run() int {
	cfg, err := config.Load(os.Getenv, true)
	log := logging.New(os.Stdout, "web", cfg.Env, cfg.Commit, cfg.LogLevel)
	if err != nil {
		log.Error("startup_failed", "error_code", "CONFIG_INVALID", "message", err.Error())
		return 1
	}
	file, err := os.OpenFile("/tmp/nginx.conf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		log.Error("startup_failed", "error_code", "CONFIG_WRITE_FAILED")
		return 1
	}
	t := template.Must(template.New("nginx").Parse(nginxTemplate))
	err = t.Execute(file, struct {
		config.Config
		ProxySeconds   int
		ConnectSeconds int
	}{cfg, int(cfg.ProxyTimeout / time.Second), int(cfg.ProxyConnectTimeout / time.Second)})
	_ = file.Close()
	if err != nil {
		log.Error("startup_failed", "error_code", "CONFIG_RENDER_FAILED")
		return 1
	}
	publicJSON, _ := json.Marshal(map[string]int{"health_timeout_ms": cfg.PublicHealthTimeoutMS})
	if err := os.WriteFile("/tmp/runtime-config.js", append(append([]byte("window.__QUIZ_CONFIG__ = "), publicJSON...), ';'), 0o600); err != nil {
		log.Error("startup_failed", "error_code", "CONFIG_WRITE_FAILED")
		return 1
	}
	cmd := exec.Command("nginx", "-c", "/tmp/nginx.conf", "-g", "daemon off;")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 1
	}
	if err := cmd.Start(); err != nil {
		log.Error("startup_failed", "error_code", "NGINX_START_FAILED")
		return 1
	}
	var wg sync.WaitGroup
	read := func(input io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			var fields map[string]any
			raw := scanner.Text()
			if raw == "" {
				continue
			}
			if json.Unmarshal([]byte(raw), &fields) != nil {
				fields = nil
			}
			if seconds, ok := fields["duration_seconds"].(float64); ok {
				fields["duration_ms"] = seconds * 1000
				delete(fields, "duration_seconds")
			}
			logging.Relay(log, fields, raw)
		}
		if scanner.Err() != nil {
			log.Error("log_read_failed", "error_code", "LOG_READ_FAILED")
		}
	}
	wg.Add(2)
	go read(stdout)
	go read(stderr)
	ended := make(chan error, 1)
	go func() { ended <- cmd.Wait() }()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signals)
	log.Info("server_starting", "port", cfg.WebPort)
	result := 0
	select {
	case err := <-ended:
		if err != nil {
			log.Error("server_failed", "error_code", "NGINX_EXIT")
			result = 1
		}
	case <-signals:
		log.Info("shutdown_started")
		_ = cmd.Process.Signal(syscall.SIGQUIT)
		select {
		case err := <-ended:
			if err != nil {
				result = 1
			}
		case <-time.After(cfg.ShutdownTimeout):
			_ = cmd.Process.Kill()
			<-ended
			log.Error("shutdown_failed", "error_code", "SHUTDOWN_TIMEOUT")
			result = 1
		}
		log.Info("shutdown_complete")
	}
	wg.Wait()
	return result
}
