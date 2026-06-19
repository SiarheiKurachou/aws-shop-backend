package main

import (
	"bufio"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	requestrouter "aws-shop-backend/src/bff-service/request-router"
)

func main() {
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %v", err)
	}

	port := os.Getenv("BFF_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", requestrouter.HandleProxyRequest)

	log.Printf("BFF service server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
