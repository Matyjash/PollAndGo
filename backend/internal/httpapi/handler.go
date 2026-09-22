package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Matyjash/PollAndGo/backend/internal/poll"
)

type Handler struct {
	repo poll.Repository
	log  *slog.Logger
}

func New(repo poll.Repository, log *slog.Logger) http.Handler {
	h := &Handler{repo: repo, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/polls", h.createPoll)
	mux.HandleFunc("GET /api/polls", h.listPolls)
	mux.HandleFunc("GET /api/polls/{id}", h.getPoll)
	mux.HandleFunc("POST /api/polls/{id}/votes", h.vote)
	return recoverMiddleware(log, mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createPoll(w http.ResponseWriter, r *http.Request) {
	var input poll.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.CreatorID = strings.TrimSpace(input.CreatorID)
	for i := range input.Options {
		input.Options[i] = strings.TrimSpace(input.Options[i])
	}
	if message := validatePoll(input); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}
	created, err := h.repo.Create(r.Context(), input)
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listPolls(w http.ResponseWriter, r *http.Request) {
	voterID := strings.TrimSpace(r.URL.Query().Get("voterId"))
	if voterID == "" || len(voterID) > 100 {
		writeError(w, http.StatusBadRequest, "a valid voterId is required")
		return
	}
	results, err := h.repo.ListForUser(r.Context(), voterID)
	if err != nil {
		h.internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) getPoll(w http.ResponseWriter, r *http.Request) {
	result, err := h.repo.Get(r.Context(), r.PathValue("id"), strings.TrimSpace(r.URL.Query().Get("voterId")))
	if err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	var input struct {
		OptionID string `json:"optionId"`
		VoterID  string `json:"voterId"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	input.OptionID, input.VoterID = strings.TrimSpace(input.OptionID), strings.TrimSpace(input.VoterID)
	if input.OptionID == "" || input.VoterID == "" || len(input.VoterID) > 100 {
		writeError(w, http.StatusBadRequest, "optionId and a valid voterId are required")
		return
	}
	result, err := h.repo.Vote(r.Context(), r.PathValue("id"), input.OptionID, input.VoterID)
	if err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func validatePoll(input poll.CreateInput) string {
	if input.Title == "" || len(input.Title) > 200 {
		return "title must contain between 1 and 200 characters"
	}
	if input.CreatorID == "" || len(input.CreatorID) > 100 {
		return "a valid creatorId is required"
	}
	if len(input.Options) < 2 || len(input.Options) > 20 {
		return "a poll must contain between 2 and 20 options"
	}
	seen := make(map[string]bool)
	for _, option := range input.Options {
		if option == "" || len(option) > 120 {
			return "each option must contain between 1 and 120 characters"
		}
		key := strings.ToLower(option)
		if seen[key] {
			return "options must be unique"
		}
		seen[key] = true
	}
	return ""
}

func (h *Handler) writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, poll.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, poll.ErrAlreadyVoted):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, poll.ErrInvalidOption):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.internalError(w, err)
	}
}

func (h *Handler) internalError(w http.ResponseWriter, err error) {
	h.log.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "an unexpected error occurred")
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func recoverMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("panic", "value", recovered)
				writeError(w, 500, "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
