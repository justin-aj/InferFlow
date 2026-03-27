package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ajinfrank/inferflow/internal/metrics"
	"github.com/ajinfrank/inferflow/internal/otel"
	"github.com/ajinfrank/inferflow/internal/proxy"
	"github.com/ajinfrank/inferflow/internal/router"
)

type Server struct {
	cfg     Config
	httpSrv *http.Server
	router  *router.RoundRobin
	client  *proxy.Client
	metrics *metrics.State

	stopCh    chan struct{}
	stopOnce  sync.Once
	probeDone chan struct{}
}

func New(cfg Config) (*Server, error) {
	rr := router.NewRoundRobin(cfg.Backends)
	s := &Server{
		cfg:       cfg,
		router:    rr,
		client:    proxy.NewClient(cfg.BackendRequestTimout),
		metrics:   &metrics.State{},
		stopCh:    make(chan struct{}),
		probeDone: make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
	mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)

	s.httpSrv = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	s.startProber()
	return s, nil
}

func (s *Server) Run() error {
	err := s.httpSrv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown() error {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		<-s.probeDone
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"in_flight": s.metrics.InFlight(),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if !s.router.HasHealthyBackend() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "no healthy backend"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, span := otel.StartSpan(r.Context(), "router_receive")
	defer span.End()

	var req proxy.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := validateChatRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, strategySpan := otel.StartSpan(ctx, "strategy_evaluate")
	backend, err := s.router.Pick()
	strategySpan.End()
	if err != nil {
		http.Error(w, "no healthy backend available", http.StatusServiceUnavailable)
		return
	}

	s.metrics.IncInFlight()
	defer s.metrics.DecInFlight()

	ctx, inferenceSpan := otel.StartSpan(ctx, "backend_inference")
	resp, err := s.client.SendChatCompletion(ctx, backend, req)
	inferenceSpan.End()
	if err != nil {
		backend.SetHealthy(false)
		http.Error(w, "backend request failed", http.StatusBadGateway)
		return
	}

	_, streamSpan := otel.StartSpan(ctx, "response_stream")
	defer streamSpan.End()
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) startProber() {
	go func() {
		defer close(s.probeDone)

		s.probeBackends()
		ticker := time.NewTicker(s.cfg.ProbeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.probeBackends()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *Server) probeBackends() {
	for _, backend := range s.cfg.Backends {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := s.client.HealthCheck(ctx, backend)
		cancel()
		backend.SetHealthy(err == nil)
	}
}

func validateChatRequest(req proxy.ChatCompletionRequest) error {
	if strings.TrimSpace(req.Model) == "" {
		return errors.New("model is required")
	}
	if len(req.Messages) == 0 {
		return errors.New("messages must not be empty")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
