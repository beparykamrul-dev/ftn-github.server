package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type State struct {
	mu      sync.RWMutex
	Policy  map[string]any
	Devices []map[string]any
	Usage   map[string]any
	Rules   map[string][]map[string]any
}

func main() {
	s := &State{
		Policy: map[string]any{
			"version": 1, "profile": "FAMILY",
			"dns": map[string]any{"enabled": true, "dnssec": true, "encrypted": true, "resolver_preference": "FTN"},
			"resolver_policy": map[string]any{"ftn": true, "cloudflare_family": true, "public_fallback": false},
		},
		Devices: []map[string]any{{"id": "runtime-device", "name": "Android Device", "profile": "FAMILY", "status": "online", "policy_version": 1}},
		Usage:   map[string]any{"allowed": 0, "blocked": 0, "failed": 0, "resolver": "FTN", "dnssec": true},
		Rules:   map[string][]map[string]any{"domain": []map[string]any{}, "ip": []map[string]any{}},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/api/v1/family/devices", s.devices)
	mux.HandleFunc("/api/v1/family/policies/", s.policy)
	mux.HandleFunc("/api/v1/dns/usage", s.usage)
	mux.HandleFunc("/api/v1/dns/domain-rules", s.rules("domain"))
	mux.HandleFunc("/api/v1/dns/ip-rules", s.rules("ip"))
	mux.HandleFunc("/api/v1/family/profiles", s.profiles)
	mux.HandleFunc("/api/v1/dns/health", s.dnsHealth)
	mux.HandleFunc("/api/v1/providers", s.providers)
	mux.HandleFunc("/api/v1/health/summary", s.summary)
	mux.HandleFunc("/api/v1/family/android/enroll", s.enroll)
	mux.Handle("/", http.FileServer(http.Dir("../../web")))

	addr := os.Getenv("FTN_FAMILY_GUARD_ADDR")
	if addr == "" { addr = ":8095" }
	server := &http.Server{Addr: addr, Handler: cors(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("FTN Family Guard backend listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}

func health(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"service": "ftn-family-guard", "status": "ok", "time": time.Now().UTC()}) }

func (s *State) devices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Devices)
}

func (s *State) policy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Policy)
}

func (s *State) usage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Usage)
}

func (s *State) profiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	write(w, []map[string]any{
		{"id": "FAMILY", "name": "Family"}, {"id": "CHILD", "name": "Child"},
		{"id": "TEEN", "name": "Teen"}, {"id": "CUSTOM", "name": "Custom"},
	})
}

func (s *State) dnsHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	write(w, map[string]any{
		"dnssec": "required", "resolver": "healthy", "preferred": "FTN",
		"cloudflare_family": "available-by-policy", "latency_ms": 0, "checked_at": time.Now().UTC(),
	})
}

func (s *State) providers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	write(w, []map[string]any{
		{"id": "ftn", "available": true, "enabled_by_default": true, "health": "healthy"},
		{"id": "cloudflare-family", "available": true, "enabled_by_default": false, "policy_controlled": true, "health": "available-by-policy"},
	})
}

func (s *State) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { method(w); return }
	write(w, map[string]any{"backend": "healthy", "dns": "healthy", "dnssec": "required", "resolver": "FTN-preferred", "cloudflare_family": "available-by-policy", "realtime": "ready", "android": "ready"})
}

func (s *State) enroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { method(w); return }
	var request map[string]any
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
	id := "runtime-assigned"
	if v, ok := request["device_id"].(string); ok && strings.TrimSpace(v) != "" { id = strings.TrimSpace(v) }
	s.mu.Lock()
	s.Devices = append(s.Devices, map[string]any{"id": id, "name": "Android Device", "profile": "FAMILY", "status": "online", "policy_version": s.Policy["version"]})
	s.mu.Unlock()
	write(w, map[string]any{"status": "accepted", "device_id": id, "policy_version": 1})
}

func (s *State) rules(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock(); defer s.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			write(w, s.Rules[kind])
		case http.MethodPost:
			var value map[string]any
			if err := json.NewDecoder(r.Body).Decode(&value); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
			if value["value"] == nil { http.Error(w, "value is required", http.StatusBadRequest); return }
			s.Rules[kind] = append(s.Rules[kind], value)
			write(w, value)
		default: method(w)
		}
	}
}

func cors(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	origin := os.Getenv("FTN_FAMILY_GUARD_ALLOWED_ORIGIN")
	if origin == "" { origin = "*" }
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Idempotency-Key")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
	next.ServeHTTP(w, r)
}) }

func method(w http.ResponseWriter) { http.Error(w, "method not allowed", http.StatusMethodNotAllowed) }
func write(w http.ResponseWriter, value any) { w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(value) }
