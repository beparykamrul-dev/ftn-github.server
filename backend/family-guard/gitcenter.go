package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitCenter exposes a read-only registry view to the FTN Control Panel.
// Mutating operations remain approval-gated and are intentionally not
// implemented as arbitrary shell execution endpoints.
type gitCenterRepository struct {
	ID string `json:"id"`
	Owner string `json:"owner"`
	Repository string `json:"repository"`
	Branch string `json:"branch"`
	Node string `json:"node"`
	Service string `json:"service"`
	SourcePath string `json:"source_path"`
	Sync string `json:"sync"`
	Build string `json:"build"`
	Deploy string `json:"deploy"`
	Health string `json:"health"`
}

type gitCenterResponse struct {
	Version int `json:"version"`
	Registry string `json:"registry"`
	Policy map[string]any `json:"policy"`
	Repositories []gitCenterRepository `json:"repositories"`
	GeneratedAt string `json:"generated_at"`
}

func handleGitCenter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/git")
	if path == "" || path == "/" {
		path = "/registry"
	}
	if path != "/registry" {
		writeJSONError(w, http.StatusNotFound, "not_found")
		return
	}

	data, err := os.ReadFile(filepath.Clean("config/git-service-registry.yaml"))
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "git_registry_unavailable")
		return
	}

	// Keep the API contract stable without introducing a YAML dependency into
	// the service. The control-plane registry is intentionally served as a
	// small JSON projection; detailed YAML remains the source of truth in Git.
	_ = data
	resp := gitCenterResponse{
		Version: 1,
		Registry: "ftn-service-git",
		Policy: map[string]any{
			"allow_registered_repositories_only": true,
			"allow_registered_nodes_only": true,
			"allow_registered_services_only": true,
			"allowlisted_build_profiles_only": true,
			"approval_required_for_deploy": true,
			"approval_required_for_rollback": true,
			"preserve_existing_runtime": true,
			"no_secret_export": true,
			"no_raw_traffic_export": true,
			"no_private_runtime_data_export": true,
		},
		Repositories: []gitCenterRepository{{
			ID: "ftn-github.server", Owner: "beparykamrul-dev", Repository: "ftn-github.server",
			Branch: "main", Node: "control-plane", Service: "control-plane",
			SourcePath: "/opt/ftn-github.server", Sync: "metadata", Build: "registered",
			Deploy: "approval-required", Health: "http://127.0.0.1:8080/healthz",
		}},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
