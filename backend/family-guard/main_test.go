package main

import (
    "net/http/httptest"
    "strings"
    "testing"
)

func TestPolicyRequiresDNSSEC(t *testing.T) {
    s := &State{Policy: map[string]any{"version":1}, Usage: map[string]any{}}
    req := httptest.NewRequest("POST", "/api/v1/family/policies/device", strings.NewReader(`{"version":2,"dns":{"enabled":true,"dnssec":false}}`))
    rec := httptest.NewRecorder()
    s.policy(rec, req)
    if rec.Code != 400 { t.Fatalf("expected 400, got %d", rec.Code) }
}

func TestEnrollmentIsIdempotent(t *testing.T) {
    s := &State{Policy: map[string]any{"version":1}, Devices: []map[string]any{}, Usage: map[string]any{}}
    body := `{"device_id":"android-test"}`
    for i := 0; i < 2; i++ {
        req := httptest.NewRequest("POST", "/api/v1/family/android/enroll", strings.NewReader(body))
        rec := httptest.NewRecorder()
        s.enroll(rec, req)
        if rec.Code != 200 { t.Fatalf("expected 200, got %d", rec.Code) }
    }
    if len(s.Devices) != 1 { t.Fatalf("expected one device, got %d", len(s.Devices)) }
}

func TestResolverIsPolicyControlled(t *testing.T) {
    s := &State{}
    req := httptest.NewRequest("GET", "/api/v1/providers", nil)
    rec := httptest.NewRecorder()
    s.providers(rec, req)
    if rec.Code != 200 { t.Fatalf("expected 200, got %d", rec.Code) }
    if !strings.Contains(rec.Body.String(), "available-by-policy") { t.Fatal("expected policy-controlled resolver state") }
}
