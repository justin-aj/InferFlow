package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"inferflow/internal/adapter"
	"inferflow/internal/llm"
)

func main() {
	addr := getenv("LLAMA_ADAPTER_ADDR", ":9000")
	llamaURL := getenv("LLAMA_BASE_URL", "http://localhost:8080")
	modelName := getenv("LLAMA_MODEL_NAME", "Qwen/Qwen2.5-0.5B-Instruct")
	timeout := durationFromEnv("LLAMA_TIMEOUT", 60*time.Second)

	client := llm.NewClient(llamaURL, modelName, timeout)
	server := &http.Server{
		Addr:    addr,
		Handler: adapter.NewHandler(client),
	}

	log.Printf("llama adapter listening on %s, targeting %s model %s", addr, llamaURL, modelName)
	log.Fatal(server.ListenAndServe())
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
