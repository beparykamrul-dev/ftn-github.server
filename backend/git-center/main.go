package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Service struct {
	ID, Owner, Repository, Branch, Node, Service, SourcePath, Sync, Build, Deploy, Health string
}

type AuditEntry struct {
	IDempotencyKey string    `json:"idempotency_key"`
	Action         string    `json:"action"`
	ServiceID      string    `json:"service_id"`
	Approval       string    `json:"approval"`
	Status         string    `json:"status"`
	Executed       bool      `json:"executed"`
	Health         string    `json:"health,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type State struct {
	Services []Service
	mu       sync.Mutex
	Audit    map[string]AuditEntry
}

var healthClient = &http.Client{Timeout: 5 * time.Second}

func main() {
	s := &State{Services: []Service{{ID: "ftn-github.server", Owner: "beparykamrul-dev", Repository: "ftn-github.server", Branch: "main", Node: "control-plane", Service: "control-plane", SourcePath: "/opt/ftn-github.server", Sync: "metadata", Build: "registered", Deploy: "approval-required", Health: "http://127.0.0.1:8080/healthz"}}, Audit: map[string]AuditEntry{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"service": "ftn-git-center", "status": "ok", "time": time.Now().UTC()}) })
	mux.HandleFunc("/api/v1/git/services", s.services)
	mux.HandleFunc("/api/v1/git/sync", s.action("sync"))
	mux.HandleFunc("/api/v1/git/build", s.action("build"))
	mux.HandleFunc("/api/v1/git/health", s.health)
	mux.HandleFunc("/api/v1/git/deploy", s.approval("deploy"))
	mux.HandleFunc("/api/v1/git/rollback", s.approval("rollback"))
	mux.HandleFunc("/api/v1/git/execute", s.execute)
	mux.HandleFunc("/api/v1/git/audit", s.audit)
	addr := os.Getenv("FTN_GIT_CENTER_ADDR"); if addr == "" { addr = ":8096" }
	root := cors(mux)
	server := &http.Server{Addr: addr, Handler: root, ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second, WriteTimeout: 15*time.Second}
	fmt.Printf("FTN Git Center listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil { panic(err) }
}

func (s *State) services(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet { method(w); return }; write(w, map[string]any{"services": s.Services, "policy": map[string]any{"registered_only": true, "allowlisted_build_profiles_only": true, "deploy_approval_required": true, "rollback_approval_required": true, "executor_enabled": os.Getenv("FTN_GIT_EXECUTOR_ENABLED") == "true"}}) }

func (s *State) action(kind string) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { method(w); return }
	var q struct{ ID string `json:"id"` }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&q); err != nil || strings.TrimSpace(q.ID) == "" { http.Error(w, "registered service id is required", 400); return }
	for _, svc := range s.Services { if svc.ID != q.ID { continue }; if kind == "sync" { write(w, map[string]any{"action":"sync","status":"accepted","mode":svc.Sync,"repository":svc.Repository,"branch":svc.Branch}); return }; if svc.Build != "registered" { http.Error(w,"build profile is not allowlisted",403); return }; write(w,map[string]any{"action":"build","status":"accepted","build_profile":svc.Build,"repository":svc.Repository,"branch":svc.Branch}); return }
	http.Error(w,"registered service not found",404)
} }

func (s *State) health(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet { method(w); return }; out:=[]map[string]any{}; for _,svc:=range s.Services { out=append(out,s.checkHealth(svc)) }; write(w,map[string]any{"services":out}) }

func (s *State) checkHealth(svc Service) map[string]any { item:=map[string]any{"id":svc.ID,"service":svc.Service,"node":svc.Node,"health_url":svc.Health,"status":"unknown"}; if svc.Health!="" { b,err:=healthClient.Get(svc.Health); if err==nil { item["http_status"]=b.StatusCode; if b.StatusCode >= 200 && b.StatusCode < 400 { item["status"]="healthy" } else { item["status"]="unhealthy" }; b.Body.Close() } else { item["status"]="unreachable" } }; if svc.SourcePath!="" { if h,err:=gitHead(svc.SourcePath); err==nil { item["head"]=h } }; return item }

func (s *State) approval(kind string) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodPost { method(w); return }; var q struct{ ID string `json:"id"`; Approval string `json:"approval"`; IdempotencyKey string `json:"idempotency_key"` }; if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)).Decode(&q); err!=nil { http.Error(w,"invalid json",400); return }; q.ID=strings.TrimSpace(q.ID); q.Approval=strings.ToLower(strings.TrimSpace(q.Approval)); q.IdempotencyKey=strings.TrimSpace(q.IdempotencyKey); if q.ID=="" { http.Error(w,"registered service id is required",400); return }; if !s.registered(q.ID) { http.Error(w,"registered service not found",404); return }; if q.Approval != "approved" { write(w,map[string]any{"action":kind,"status":"approval-required","executed":false,"service_id":q.ID}); return }; if q.IdempotencyKey=="" { http.Error(w,"idempotency_key is required for approved execution",400); return }; s.mu.Lock(); if prior,ok:=s.Audit[q.IdempotencyKey]; ok { s.mu.Unlock(); write(w,prior); return }; entry:=AuditEntry{IdempotencyKey:q.IdempotencyKey,Action:kind,ServiceID:q.ID,Approval:q.Approval,Status:"approved",Executed:false,CreatedAt:time.Now().UTC()}; s.Audit[q.IdempotencyKey]=entry; s.mu.Unlock(); write(w,entry) } }

func (s *State) execute(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodPost { method(w); return }; if os.Getenv("FTN_GIT_EXECUTOR_ENABLED") != "true" { http.Error(w,"runtime executor is disabled",503); return }; var q struct{ Action string `json:"action"`; ID string `json:"id"`; IdempotencyKey string `json:"idempotency_key"` }; if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)).Decode(&q); err!=nil { http.Error(w,"invalid json",400); return }; q.Action=strings.ToLower(strings.TrimSpace(q.Action)); q.ID=strings.TrimSpace(q.ID); q.IdempotencyKey=strings.TrimSpace(q.IdempotencyKey); if q.IdempotencyKey=="" { http.Error(w,"idempotency_key is required",400); return }; s.mu.Lock(); entry,ok:=s.Audit[q.IdempotencyKey]; s.mu.Unlock(); if !ok || entry.ServiceID!=q.ID || entry.Action!=q.Action || entry.Approval!="approved" { http.Error(w,"approved execution record not found",403); return }; svc,ok:=s.find(q.ID); if !ok { http.Error(w,"registered service not found",404); return }; result:=executeRegistered(context.Background(),svc,q.Action); if result.Executed { h:=s.checkHealth(svc); status,_:=h["status"].(string); entry.Health=status; if status=="healthy" { entry.Status="completed" } else { entry.Status="executed-unhealthy" }; entry.Executed=true } else { entry.Status=result.Status }; s.mu.Lock(); s.Audit[q.IdempotencyKey]=entry; s.mu.Unlock(); write(w,map[string]any{"audit":entry,"execution":result}) }

func (s *State) audit(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet { method(w); return }; s.mu.Lock(); defer s.mu.Unlock(); out:=make([]AuditEntry,0,len(s.Audit)); for _,entry:=range s.Audit { out=append(out,entry) }; write(w,map[string]any{"entries":out}) }
func (s *State) registered(id string) bool { _,ok:=s.find(id); return ok }
func (s *State) find(id string) (Service,bool) { for _,svc:=range s.Services { if svc.ID==id { return svc,true } }; return Service{},false }
func gitHead(path string) (string,error) { if _,err:=os.Stat(filepath.Join(path,".git")); err!=nil{return "",err}; b,err:=exec.Command("git","-C",path,"rev-parse","--short","HEAD").Output(); return strings.TrimSpace(string(b)),err }
func method(w http.ResponseWriter){http.Error(w,"method not allowed",405)}
func write(w http.ResponseWriter,v any){w.Header().Set("Content-Type","application/json");_ = json.NewEncoder(w).Encode(v)}
func cors(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ o:=os.Getenv("FTN_GIT_CENTER_ALLOWED_ORIGIN"); if o!="" { w.Header().Set("Access-Control-Allow-Origin",o); w.Header().Set("Vary","Origin") }; w.Header().Set("Access-Control-Allow-Headers","Content-Type,Authorization,Idempotency-Key"); w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS"); if r.Method==http.MethodOptions { w.WriteHeader(204); return }; next.ServeHTTP(w,r) }) }
