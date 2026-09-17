package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitCenterServicesRegisteredOnly(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1", Repository: "repo", Branch: "main", Sync: "metadata", Build: "registered"}}}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/git/services", nil)
	w := httptest.NewRecorder()
	s.services(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "registered_only") { t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String()) }
}

func TestGitCenterUnknownServiceRejected(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1", Build: "registered"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/build", strings.NewReader(`{"id":"unknown"}`))
	w := httptest.NewRecorder()
	s.action("build")(w, r)
	if w.Code != http.StatusNotFound { t.Fatalf("expected 404, got %d", w.Code) }
}

func TestGitCenterDeployRequiresApproval(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(`{"id":"svc-1","approval":"pending"}`))
	w := httptest.NewRecorder()
	s.approval("deploy")(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "approval-required") { t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String()) }
}

func TestGitCenterApprovedDeployDoesNotExecute(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(`{"id":"svc-1","approval":"approved"}`))
	w := httptest.NewRecorder()
	s.approval("deploy")(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"executed":false`) { t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String()) }
}
