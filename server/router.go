package server

import (
	"encoding/json"
	"gsm/core"
	"net/http"
	"time"
)

// Server HTTP sunucu yapısı
type Server struct {
	Engine *core.MatrixEngine
	Addr   string
}

// NewServer yeni bir HTTP sunucu oluşturur
func NewServer(addr string, engine *core.MatrixEngine) *Server {
	return &Server{
		Engine: engine,
		Addr:   addr,
	}
}

// Start HTTP sunucusunu başlatır
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Statik Frontend Dosyaları Servisi
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	// API Yolları
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/publish", s.handlePublish)

	// CORS ve Logging Middleware
	handler := s.loggingMiddleware(s.corsMiddleware(mux))

	return http.ListenAndServe(s.Addr, handler)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Konsola basit log yazdıralım
		println("[HTTP]", r.Method, r.URL.Path, "Duration:", time.Since(start).String())
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := map[string]any{
		"status":      "OPERATIONAL",
		"version":     "1.0.0-MVP",
		"timestamp":   time.Now(),
		"engine_info": "Global Scalable Matrix Core active",
	}
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	events := s.Engine.GetRecentEvents()
	json.NewEncoder(w).Encode(events)
}

type PublishRequest struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}

	event := core.Event{
		ID:        "evt_manual_" + time.Now().Format("20060102150405"),
		Type:      req.Type,
		Payload:   req.Payload,
		Timestamp: time.Now(),
	}

	s.Engine.Publish(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Event published successfully",
		"event_id": event.ID,
	})
}
