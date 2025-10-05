package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type RedisConfig struct {
	Scheme   string
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

func parseRedisURL(redisURL string) (*RedisConfig, error) {
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	// Defaults
	port := "6379"
	db := 0

	// Extract host and port
	host := parsedURL.Hostname()
	if parsedPort := parsedURL.Port(); parsedPort != "" {
		port = parsedPort
	}

	// User info (optional)
	var username, password string
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
		if p, ok := parsedURL.User.Password(); ok {
			password = p
		}
	}

	// Parse DB from path (e.g. "/0")
	if parsedURL.Path != "" {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		if path != "" {
			if dbNum, err := strconv.Atoi(path); err == nil {
				db = dbNum
			} else {
				return nil, fmt.Errorf("invalid database index: %w", err)
			}
		}
	}

	return &RedisConfig{
		Scheme:   parsedURL.Scheme,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		DB:       db,
	}, nil
}
