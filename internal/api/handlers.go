package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/patrickjmcd/ytdl-triage/internal/config"
	"github.com/patrickjmcd/ytdl-triage/internal/triage"
)

type Server struct {
	cfg config.Config
}

func NewServer(cfg config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/items", s.listItems)
	mux.HandleFunc("GET /api/items/{id}/thumbnail", s.thumbnail)
	mux.HandleFunc("GET /api/items/{id}/stream", s.stream)
	mux.HandleFunc("POST /api/items/{id}/accept", s.accept)
	mux.HandleFunc("DELETE /api/items/{id}", s.reject)
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(s.cfg.OrganizedDir); err != nil {
		http.Error(w, "organized dir not reachable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := triage.Scan(s.cfg.NeedsReviewDir)
	if err != nil {
		log.Printf("scan failed: %v", err)
		http.Error(w, "scan failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}

func (s *Server) thumbnail(w http.ResponseWriter, r *http.Request) {
	item, err := triage.FindItem(s.cfg.NeedsReviewDir, r.PathValue("id"))
	if err != nil || !item.HasThumbnail {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, item.ThumbPath())
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	item, err := triage.FindItem(s.cfg.NeedsReviewDir, r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, item.MediaPath())
}

type acceptRequest struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func (s *Server) accept(w http.ResponseWriter, r *http.Request) {
	item, err := triage.FindItem(s.cfg.NeedsReviewDir, r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var req acceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Artist == "" || req.Title == "" {
		http.Error(w, "artist and title are required", http.StatusBadRequest)
		return
	}

	dest, err := triage.Accept(item, s.cfg.OrganizedDir, s.cfg.UnknownArtist, s.cfg.UnknownTitle, req.Artist, req.Title)
	if err != nil {
		log.Printf("accept failed for %s: %v", r.PathValue("id"), err)
		http.Error(w, "failed to accept item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"path": dest})
}

func (s *Server) reject(w http.ResponseWriter, r *http.Request) {
	item, err := triage.FindItem(s.cfg.NeedsReviewDir, r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := triage.Reject(item); err != nil {
		log.Printf("reject failed for %s: %v", r.PathValue("id"), err)
		http.Error(w, "failed to delete item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json failed: %v", err)
	}
}
