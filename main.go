package main

import (
	"crypto/subtle"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen      string       `yaml:"listen"`
	ForwardTo   string       `yaml:"forward_to"`
	Credentials []Credential `yaml:"credentials"`
}

type Credential struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to YAML config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	targetURL, err := url.Parse(cfg.ForwardTo)
	if err != nil {
		log.Fatalf("invalid forward_to URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	handler := withBasicAuth(cfg.Credentials, proxy)

	log.Printf("proxy listening on %s and forwarding to %s", cfg.Listen, targetURL.String())
	if err := http.ListenAndServe(cfg.Listen, handler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.ForwardTo == "" {
		return nil, errors.New("forward_to is required")
	}
	if len(cfg.Credentials) == 0 {
		return nil, errors.New("at least one credential is required")
	}
	for i, cred := range cfg.Credentials {
		if cred.Username == "" || cred.Password == "" {
			return nil, fmt.Errorf("credential at index %d must include username and password", i)
		}
	}

	return &cfg, nil
}

func withBasicAuth(credentials []Credential, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !isAuthorized(credentials, user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="proxy"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthorized(credentials []Credential, username, password string) bool {
	for _, cred := range credentials {
		if subtle.ConstantTimeCompare([]byte(cred.Username), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(cred.Password), []byte(password)) == 1 {
			return true
		}
	}
	return false
}
