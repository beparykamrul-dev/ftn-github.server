package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Service struct {
	ID, Owner, Repository, Branch, Node, Service, SourcePath, Sync, Build, Deploy, Health string
}

type State struct{ Services []Service }

func main() {
	s := &State{Services: []Service{{ID: "ftn-github.server", Owner: "beparykamrul-dev", Repository: "ftn-github.server", Branch: "main", Node: "control-plane", Service: "control-plane", SourcePath: "/opt/ftn-github.server", Sync: "metadata", Build: "registered", Deploy: "approval-required", Health: "http://127.0.0.1:8080/healthz"}}}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { write(w, map[string]any{"service": "ftn-git-center", "status": "ok", "time": time.Now().UTC()}) })
	mux.HandleFunc("/api/v1/git/services", s.services)
	mux.HandleFunc("/api/v1/git/sync", s.action("sync"))
	mux.HandleFunc("/api/v1/git/build", s.action("build"))
	mux.HandleFunc("/api/v1/git/health", s.health)
	mux.HandleFunc("/api/v1/git/deploy", s.approval("deploy"))
	mux.HandleFunc("/api/v1/git/rollback", s.approval("rollback"))
	addr := os.Getenv("FTN_GIT_CENTER_ADDR"); if addr == "" { addr = ":8096" }
	root := cors(mux)
	server := &http.Server{Addr: addr, Handler: root, ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second, WriteTimeout: 15*time.Second}
	fmt.Printf("FTN Git Center listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil { panic(err) }
}

func (s *State) services(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet { method(w); return }; write(w, map[string]any{"services": s.Services, "policy": map[string]any{"registered_only": true, "allowlisted_build_profiles_only": true, "deploy_approval_required": true, "rollback_approval_required": true}}) }

func (s *State) action(kind string) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { method(w); return }
	var q struct{ ID string `json:"id"` }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&q); err != nil || strings.TrimSpace(q.ID) == "" { http.Error(w, "registered service id is required", 400); return }
	for _, svc := range s.Services { if svc.ID != q.ID { continue }; if kind == "sync" { write(w, map[string]any{"action":"sync","status":"accepted","mode":svc.Sync,"repository":svc.Repository,"branch":svc.Branch}); return }; if svc.Build != "registered" { http.Error(w,"build profile is not allowlisted",403); return }; write(w,map[string]any{"action":"build","status":"accepted","build_profile":svc.Build,"repository":svc.Repository,"branch":svc.Branch}); return }
	http.Error(w,"registered service not found",404)
} }

func (s *State) health(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodGet { method(w); return }; out:=[]map[string]any{}; for _,svc:=range s.Services { item:=map[string]any{"id":svc.ID,"service":svc.Service,"node":svc.Node,"health_url":svc.Health,"status":"unknown"}; if svc.Health!="" { if b,err:=http.Get(svc.Health); err==nil { item["status"]="healthy"; item["http_status"]=b.StatusCode; b.Body.Close() } else { item["status"]="unreachable" } }; if svc.SourcePath!="" { if h,err:=gitHead(svc.SourcePath); err==nil { item["head"]=h } }; out=append(out,item) }; write(w,map[string]any{"services":out}) }

func (s *State) approval(kind string) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) { if r.Method != http.MethodPost { method(w); return }; var q struct{ ID string `json:"id"`; Approval string `json:"approval"` }; if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)).Decode(&q); err!=nil { http.Error(w,"invalid json",400); return }; if strings.TrimSpace(q.ID)=="" { http.Error(w,"registered service id is required",400); return }; if strings.ToLower(strings.TrimSpace(q.Approval)) != "approved" { write(w,map[string]any{"action":kind,"status":"approval-required","executed":false,"service_id":q.ID}); return }; write(w,map[string]any{"action":kind,"status":"approval-recorded","executed":false,"service_id":q.ID,"note":"Execution is intentionally disabled in Git Center API; registered runtime automation must perform the approved action."}) } }

func gitHead(path string) (string,error) { if _,err:=os.Stat(filepath.Join(path,".git")); err!=nil{return "",err}; b,err:=exec.Command("git","-C",path,"rev-parse","--short","HEAD").Output(); return strings.TrimSpace(string(b)),err }
func method(w http.ResponseWriter){http.Error(w,"method not allowed",405)}
func write(w http.ResponseWriter,v any){w.Header().Set("Content-Type","application/json");_ = json.NewEncoder(w).Encode(v)}
func cors(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ o:=os.Getenv("FTN_GIT_CENTER_ALLOWED_ORIGIN"); if o!="" { w.Header().Set("Access-Control-Allow-Origin",o); w.Header().Set("Vary","Origin") }; w.Header().Set("Access-Control-Allow-Headers","Content-Type,Authorization,Idempotency-Key"); w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS"); if r.Method==http.MethodOptions { w.WriteHeader(204); return }; next.ServeHTTP(w,r) }) }
