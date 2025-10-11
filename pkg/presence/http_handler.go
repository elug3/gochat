package presence

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Handler struct {
	mux       *http.ServeMux
	Presences *PresenceService
}

func NewHandler(opts *HttpOptions) (*Handler, error) {
	presences, err := NewPresenceService(opts)
	if err != nil {
		return nil, err
	}
	h := Handler{
		mux:       http.NewServeMux(),
		Presences: presences,
	}

	h.registerRoutes()

	return &h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/users/{userId}/presence", func(w http.ResponseWriter, r *http.Request) {
		userId, err := parseUserId(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		presence, err := h.Presences.GetPresence(userId)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSONResponse(w, http.StatusOK, presence)
	})
}

func parseUserId(r *http.Request) (int32, error) {
	uidStr := r.PathValue("userId")
	if uidStr == "" {
		return 0, fmt.Errorf("userId is required")
	}
	uid, err := strconv.ParseInt(uidStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid userId: %v", err)
	}
	return int32(uid), nil
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
