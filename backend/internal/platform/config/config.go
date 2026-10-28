package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	PublicHealthTimeoutMS                                                                      int
	Env, Name, Version, Commit, LogLevel, ListenHost, APIHost                                  string
	APIPort, WebPort                                                                           int
	ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout, ProxyTimeout, ProxyConnectTimeout time.Duration
}

var hostPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,252}$`)

// Load has no implicit defaults: the committed .env.example is the config template.
func Load(get func(string) string, web bool) (Config, error) {
	var c Config
	for _, field := range []struct {
		key string
		dst *string
	}{
		{"APP_ENV", &c.Env}, {"APP_NAME", &c.Name}, {"APP_VERSION", &c.Version},
		{"BUILD_COMMIT", &c.Commit}, {"LOG_LEVEL", &c.LogLevel},
	} {
		*field.dst = get(field.key)
		if strings.TrimSpace(*field.dst) == "" || strings.ContainsAny(*field.dst, "\r\n") {
			return c, fmt.Errorf("%s is required and must be one line", field.key)
		}
	}
	if c.Env != "dev" && c.Env != "test" && c.Env != "demo" {
		return c, fmt.Errorf("APP_ENV must be dev, test or demo")
	}
	if c.LogLevel != "DEBUG" && c.LogLevel != "INFO" && c.LogLevel != "WARN" && c.LogLevel != "ERROR" {
		return c, fmt.Errorf("LOG_LEVEL is invalid")
	}
	var err error
	if c.APIPort, err = number(get, "API_PORT", 1, 65535); err != nil {
		return c, err
	}
	if web {
		c.APIHost = get("API_HOST")
		if !hostPattern.MatchString(c.APIHost) {
			return c, fmt.Errorf("API_HOST must be a hostname")
		}
		if c.WebPort, err = number(get, "WEB_LISTEN_PORT", 1024, 65535); err != nil {
			return c, err
		}
		if c.ProxyTimeout, err = seconds(get, "PROXY_TIMEOUT_SECONDS", 1, 30); err != nil {
			return c, err
		}
		if c.ProxyConnectTimeout, err = seconds(get, "PROXY_CONNECT_TIMEOUT_SECONDS", 1, 5); err != nil {
			return c, err
		}
		if c.PublicHealthTimeoutMS, err = number(get, "PUBLIC_HEALTH_TIMEOUT_MS", 100, 30000); err != nil {
			return c, err
		}
	} else {
		c.ListenHost = get("API_LISTEN_HOST")
		if net.ParseIP(c.ListenHost) == nil {
			return c, fmt.Errorf("API_LISTEN_HOST must be an IP address")
		}
		for _, field := range []struct {
			key string
			dst *time.Duration
			max int
		}{
			{"HTTP_READ_TIMEOUT_SECONDS", &c.ReadTimeout, 60},
			{"HTTP_WRITE_TIMEOUT_SECONDS", &c.WriteTimeout, 60},
			{"HTTP_IDLE_TIMEOUT_SECONDS", &c.IdleTimeout, 300},
		} {
			*field.dst, err = seconds(get, field.key, 1, field.max)
			if err != nil {
				return c, err
			}
		}
	}
	c.ShutdownTimeout, err = seconds(get, "SHUTDOWN_TIMEOUT_SECONDS", 1, 20)
	return c, err
}

func number(get func(string) string, key string, min, max int) (int, error) {
	n, err := strconv.Atoi(get(key))
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be an integer in [%d,%d]", key, min, max)
	}
	return n, nil
}

func seconds(get func(string) string, key string, min, max int) (time.Duration, error) {
	n, err := number(get, key, min, max)
	return time.Duration(n) * time.Second, err
}
