package main

import (
	"fmt"
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

func TestGitCenterApprovedDeployRequiresIdempotency(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(`{"id":"svc-1","approval":"approved"}`))
	w := httptest.NewRecorder()
	s.approval("deploy")(w, r)
	if w.Code != http.StatusBadRequest { t.Fatalf("expected 400, got %d", w.Code) }
}

func TestGitCenterApprovedDeployIsIdempotent(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	body := `{"id":"svc-1","approval":"approved","idempotency_key":"deploy-1"}`
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(body))
		w := httptest.NewRecorder()
		s.approval("deploy")(w, r)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"idempotency_key":"deploy-1"`) { t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String()) }
	}
	if len(s.Audit) != 1 { t.Fatalf("expected one audit record, got %d", len(s.Audit)) }
}

func TestGitCenterApprovedDeployDoesNotExecuteByApprovalAlone(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(`{"id":"svc-1","approval":"approved","idempotency_key":"deploy-2"}`))
	w := httptest.NewRecorder()
	s.approval("deploy")(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"executed":false`) { t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String()) }
}

func TestGitCenterUnknownApprovalRejected(t *testing.T) {
	s := &State{Services: []Service{{ID: "svc-1"}}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/git/deploy", strings.NewReader(`{"id":"unknown","approval":"approved","idempotency_key":"x"}`))
	w := httptest.NewRecorder()
	s.approval("deploy")(w, r)
	if w.Code != http.StatusNotFound { t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String()) }
}

func TestGitCenterHealthNon2xxIsUnhealthy(t *testing.T) {
	old := healthClient
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "bad", http.StatusInternalServerError) }))
	defer ts.Close()
	healthClient = ts.Client()
	defer func() { healthClient = old }()

	s := &State{Services: []Service{{ID: "svc-1", Service: "test", Health: ts.URL}}}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/git/health", nil)
	w := httptest.NewRecorder()
	s.health(w, r)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `"status":"unhealthy"`) || !strings.Contains(body, fmt.Sprintf(`"http_status":%d`, http.StatusInternalServerError)) {
		t.Fatalf("unexpected response: %d %s", w.Code, body)
	}
}
