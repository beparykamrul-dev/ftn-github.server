package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitCenterRegistry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/registry", nil)
	rec := httptest.NewRecorder()
	handleGitCenter(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGitCenterRejectsMutation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/registry", nil)
	rec := httptest.NewRecorder()
	handleGitCenter(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
