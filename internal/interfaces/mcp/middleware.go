package mcp

import (
	"encoding/json"
	"net/http"
)

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	if s.licenseHandler != nil {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/license/hwid":
			s.licenseHandler.HWID(w, r)
			return
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/license/status":
			s.licenseHandler.Status(w, r)
			return
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/license/check":
			s.licenseHandler.CheckFeature(w, r)
			return
		}
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "application/json required", http.StatusBadRequest)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, _ := s.HandleRequest(r.Context(), req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}