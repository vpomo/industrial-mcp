package rest

import (
	"encoding/json"
	"net/http"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/service"
)

type ExposedTagHandler struct {
	svc *service.ExposedTagService
}

func NewExposedTagHandler(svc *service.ExposedTagService) *ExposedTagHandler {
	return &ExposedTagHandler{svc: svc}
}

type CreateExposedTagRequest struct {
	Name         string `json:"name"`
	DataSourceID string `json:"data_source_id"`
	NodeID       string `json:"node_id"`
	DataType     string `json:"data_type"`
	ReadOnly     bool   `json:"read_only"`
}

func (h *ExposedTagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateExposedTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tag, err := h.svc.Create(r.Context(), req.Name, req.DataSourceID, req.NodeID, req.DataType, req.ReadOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (h *ExposedTagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *ExposedTagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	tag, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (h *ExposedTagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExposedTagHandler) ReadValue(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	tag, err := h.svc.ReadValue(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (h *ExposedTagHandler) WriteValue(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.WriteValue(r.Context(), id, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}