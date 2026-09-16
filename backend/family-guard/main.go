package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "sync"
    "time"
)

type State struct {
    mu sync.RWMutex
    Policy map[string]any
    Devices []map[string]any
    Usage map[string]any
    Rules map[string][]map[string]any
}

func main() {
    s := &State{
        Policy: map[string]any{"version": 1, "profile": "FAMILY", "dns": map[string]any{"enabled": true, "dnssec": true, "encrypted": true}},
        Devices: []map[string]any{{"id":"runtime-device","name":"Android Device","profile":"FAMILY","status":"online","policy_version":1}},
        Usage: map[string]any{"allowed":0,"blocked":0,"failed":0,"resolver":"FTN","dnssec":true},
        Rules: map[string][]map[string]any{"domain":{},"ip":{}},
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
    addr := os.Getenv("FTN_FAMILY_GUARD_ADDR"); if addr == "" { addr = ":8095" }
    srv := &http.Server{Addr: addr, Handler: cors(mux), ReadHeaderTimeout: 5*time.Second}
    log.Printf("FTN Family Guard backend listening on %s", srv.Addr)
    log.Fatal(srv.ListenAndServe())
}
func health(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"service":"ftn-family-guard","status":"ok"}) }
func (s *State) devices(w http.ResponseWriter, _ *http.Request) { s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Devices) }
func (s *State) policy(w http.ResponseWriter, _ *http.Request) { s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Policy) }
func (s *State) usage(w http.ResponseWriter, _ *http.Request) { s.mu.RLock(); defer s.mu.RUnlock(); write(w, s.Usage) }
func (s *State) profiles(w http.ResponseWriter, _ *http.Request) { write(w, []string{"FAMILY","CHILD","TEEN","CUSTOM"}) }
func (s *State) dnsHealth(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"dnssec":"healthy","resolver":"healthy","latency_ms":0,"checked_at":time.Now().UTC()}) }
func (s *State) providers(w http.ResponseWriter, _ *http.Request) { write(w, []map[string]any{{"id":"ftn","enabled":true,"health":"healthy"},{"id":"cloudflare-family","enabled":false,"health":"policy-controlled"}}) }
func (s *State) summary(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"backend":"healthy","dns":"healthy","dnssec":"required","realtime":"ready","android":"ready"}) }
func (s *State) enroll(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodPost { http.Error(w,"method not allowed",405); return }; write(w,map[string]any{"status":"accepted","device_id":"runtime-assigned","policy_version":1}) }
func (s *State) rules(kind string) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) { s.mu.Lock(); defer s.mu.Unlock(); if r.Method==http.MethodPost { var v map[string]any; if json.NewDecoder(r.Body).Decode(&v)==nil { s.Rules[kind]=append(s.Rules[kind],v) } }; write(w,s.Rules[kind]) } }
func cors(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ w.Header().Set("Access-Control-Allow-Origin","*"); w.Header().Set("Access-Control-Allow-Headers","Content-Type,Authorization,Idempotency-Key"); w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS"); if r.Method=="OPTIONS" { w.WriteHeader(204); return }; next.ServeHTTP(w,r) }) }
func write(w http.ResponseWriter, v any) { w.Header().Set("Content-Type","application/json"); _=json.NewEncoder(w).Encode(v) }
