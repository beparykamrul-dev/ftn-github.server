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

type State struct { mu sync.RWMutex; Policy map[string]any; Devices []map[string]any; Usage map[string]any; Rules map[string][]map[string]any }

func main() {
    s := &State{
        Policy: map[string]any{"version":1,"profile":"FAMILY","dns":map[string]any{"enabled":true,"dnssec":true,"encrypted":true,"resolver_preference":"FTN"},"resolver_policy":map[string]any{"ftn":true,"cloudflare_family":true,"public_fallback":false}},
        Devices: []map[string]any{{"id":"runtime-device","name":"Android Device","profile":"FAMILY","status":"online","policy_version":1}},
        Usage: map[string]any{"allowed":0,"blocked":0,"failed":0,"resolver":"FTN","dnssec":true,"policy_version":1},
        Rules: map[string][]map[string]any{"domain":{},"ip":{}},
    }
    mux:=http.NewServeMux()
    mux.HandleFunc("/healthz",health); mux.HandleFunc("/api/v1/family/devices",s.devices); mux.HandleFunc("/api/v1/family/policies/",s.policy)
    mux.HandleFunc("/api/v1/dns/usage",s.usage); mux.HandleFunc("/api/v1/dns/domain-rules",s.rules("domain")); mux.HandleFunc("/api/v1/dns/ip-rules",s.rules("ip"))
    mux.HandleFunc("/api/v1/family/profiles",s.profiles); mux.HandleFunc("/api/v1/dns/health",s.dnsHealth); mux.HandleFunc("/api/v1/providers",s.providers)
    mux.HandleFunc("/api/v1/health/summary",s.summary); mux.HandleFunc("/api/v1/family/android/enroll",s.enroll); mux.Handle("/",http.FileServer(http.Dir("../../web")))
    addr:=os.Getenv("FTN_FAMILY_GUARD_ADDR"); if addr=="" {addr=":8095"}
    srv:=&http.Server{Addr:addr,Handler:cors(mux),ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:15*time.Second,IdleTimeout:60*time.Second}
    log.Printf("FTN Family Guard backend listening on %s",addr); log.Fatal(srv.ListenAndServe())
}
func health(w http.ResponseWriter,_ *http.Request){write(w,map[string]any{"service":"ftn-family-guard","status":"ok","time":time.Now().UTC()})}
func(s *State)devices(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};s.mu.RLock();defer s.mu.RUnlock();write(w,s.Devices)}
func(s *State)policy(w http.ResponseWriter,r *http.Request){
    switch r.Method {
    case http.MethodGet: s.mu.RLock();defer s.mu.RUnlock();write(w,s.Policy)
    case http.MethodPost:
        var next map[string]any; if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,64<<10)).Decode(&next);err!=nil{http.Error(w,"invalid json",400);return}
        version,ok:=next["version"].(float64);if !ok||version<1{http.Error(w,"version is required",400);return}
        dns,ok:=next["dns"].(map[string]any);if !ok{http.Error(w,"dns policy is required",400);return}
        dnssec,ok:=dns["dnssec"].(bool);if !ok||!dnssec{http.Error(w,"dnssec must remain required",400);return}
        next["resolver_policy"]=map[string]any{"ftn":true,"cloudflare_family":true,"public_fallback":false}
        s.mu.Lock();s.Policy=next;s.Usage["policy_version"]=version;s.mu.Unlock();write(w,next)
    default:method(w)
    }
}
func(s *State)usage(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};s.mu.RLock();defer s.mu.RUnlock();write(w,s.Usage)}
func(s *State)profiles(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};write(w,[]map[string]any{{"id":"FAMILY","name":"Family"},{"id":"CHILD","name":"Child"},{"id":"TEEN","name":"Teen"},{"id":"CUSTOM","name":"Custom"}})}
func(s *State)dnsHealth(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};write(w,map[string]any{"dnssec":"required","resolver":"healthy","preferred":"FTN","cloudflare_family":"available-by-policy","latency_ms":0,"checked_at":time.Now().UTC()})}
func(s *State)providers(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};write(w,[]map[string]any{{"id":"ftn","available":true,"enabled_by_default":true,"health":"healthy"},{"id":"cloudflare-family","available":true,"enabled_by_default":false,"policy_controlled":true,"health":"available-by-policy"}})}
func(s *State)summary(w http.ResponseWriter,r *http.Request){if r.Method!=http.MethodGet{method(w);return};write(w,map[string]any{"backend":"healthy","dns":"healthy","dnssec":"required","resolver":"FTN-preferred","cloudflare_family":"available-by-policy","realtime":"ready","android":"ready"})}
func(s *State)enroll(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost{method(w);return};var q map[string]any
    if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,32<<10)).Decode(&q);err!=nil{http.Error(w,"invalid json",400);return}
    id,_:=q["device_id"].(string);id=strings.TrimSpace(id);if id==""{http.Error(w,"device_id is required",400);return}
    s.mu.Lock();defer s.mu.Unlock();for _,d:=range s.Devices{if d["id"]==id{write(w,map[string]any{"status":"already-enrolled","device_id":id,"policy_version":s.Policy["version"]});return}}
    s.Devices=append(s.Devices,map[string]any{"id":id,"name":"Android Device","profile":"FAMILY","status":"online","policy_version":s.Policy["version"]})
    write(w,map[string]any{"status":"accepted","device_id":id,"policy_version":s.Policy["version"]})
}
func(s *State)rules(kind string)http.HandlerFunc{return func(w http.ResponseWriter,r *http.Request){
    switch r.Method{case http.MethodGet:s.mu.RLock();defer s.mu.RUnlock();write(w,s.Rules[kind]);case http.MethodPost:
        var v map[string]any;if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10)).Decode(&v);err!=nil{http.Error(w,"invalid json",400);return}
        value,ok:=v["value"].(string);if !ok||strings.TrimSpace(value)==""{http.Error(w,"value is required",400);return};action,ok:=v["action"].(string);if !ok||(action!="allow"&&action!="block"){http.Error(w,"action must be allow or block",400);return}
        v["value"]=strings.TrimSpace(value);s.mu.Lock();s.Rules[kind]=append(s.Rules[kind],v);s.mu.Unlock();write(w,v)
    default:method(w)}}}
func cors(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){origin:=os.Getenv("FTN_FAMILY_GUARD_ALLOWED_ORIGIN");if origin==""{origin="*"};w.Header().Set("Access-Control-Allow-Origin",origin);w.Header().Set("Vary","Origin");w.Header().Set("Access-Control-Allow-Headers","Content-Type,Authorization,Idempotency-Key");w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS");if r.Method=="OPTIONS"{w.WriteHeader(204);return};next.ServeHTTP(w,r)})}
func method(w http.ResponseWriter){http.Error(w,"method not allowed",405)}
func write(w http.ResponseWriter,v any){w.Header().Set("Content-Type","application/json");_ = json.NewEncoder(w).Encode(v)}
