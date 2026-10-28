package config

import (
	"strings"
	"testing"
)

func validEnv() map[string]string {
	return map[string]string{"APP_ENV": "test", "APP_NAME": "quiz", "APP_VERSION": "test", "BUILD_COMMIT": "test", "LOG_LEVEL": "INFO", "API_PORT": "8081", "API_LISTEN_HOST": "0.0.0.0", "API_HOST": "api", "WEB_LISTEN_PORT": "8080", "HTTP_READ_TIMEOUT_SECONDS": "5", "HTTP_WRITE_TIMEOUT_SECONDS": "10", "HTTP_IDLE_TIMEOUT_SECONDS": "60", "SHUTDOWN_TIMEOUT_SECONDS": "15", "PROXY_TIMEOUT_SECONDS": "5", "PROXY_CONNECT_TIMEOUT_SECONDS": "1", "PUBLIC_HEALTH_TIMEOUT_MS": "5000"}
}

func TestNormalExternalConfig(t *testing.T) {
	for _, web := range []bool{false, true} {
		env := validEnv()
		cfg, err := Load(func(k string) string { return env[k] }, web)
		if err != nil || cfg.APIPort != 8081 {
			t.Fatalf("valid external config: %v", err)
		}
	}
}

func TestFailureAndBoundaryConfig(t *testing.T) {
	for _, tc := range []struct {
		key, value string
		web        bool
	}{
		{"API_PORT", "0", false}, {"API_PORT", "70000", false}, {"API_PORT", "1.5", false},
		{"APP_ENV", "production-typo", false}, {"LOG_LEVEL", "verbose", false},
		{"APP_NAME", "", false}, {"HTTP_READ_TIMEOUT_SECONDS", "", false},
		{"API_HOST", "api; injected", true}, {"WEB_LISTEN_PORT", "80", true},
		{"PUBLIC_HEALTH_TIMEOUT_MS", "0", true},
		{"API_LISTEN_HOST", "", false}, {"PROXY_TIMEOUT_SECONDS", "31", true},
	} {
		t.Run(tc.key+tc.value, func(t *testing.T) {
			env := validEnv()
			env[tc.key] = tc.value
			_, err := Load(func(k string) string { return env[k] }, tc.web)
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("invalid %s must fail", tc.key)
			}
		})
	}
	for _, port := range []string{"1", "65535"} {
		env := validEnv()
		env["API_PORT"] = port
		if _, err := Load(func(k string) string { return env[k] }, false); err != nil {
			t.Fatal(err)
		}
	}
}
