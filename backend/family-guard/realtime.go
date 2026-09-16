package main

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "strings"
    "sync"
    "time"

    "nhooyr.io/websocket"
)

type RealtimeHub struct {
    mu sync.RWMutex
    conns map[string]map[*websocket.Conn]struct{}
}

func NewRealtimeHub() *RealtimeHub { return &RealtimeHub{conns: make(map[string]map[*websocket.Conn]struct{})} }

func (h *RealtimeHub) HandleWS(w http.ResponseWriter, r *http.Request) {
    deviceID := strings.TrimSpace(r.URL.Query().Get("device_id"))
    session := strings.TrimSpace(r.URL.Query().Get("session"))
    if deviceID == "" || session == "" { http.Error(w, "device_id and session are required", http.StatusBadRequest); return }
    if !validSession(deviceID, session) { http.Error(w, "invalid session", http.StatusUnauthorized); return }
    c, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled})
    if err != nil { return }
    h.mu.Lock(); if h.conns[deviceID] == nil { h.conns[deviceID] = make(map[*websocket.Conn]struct{}) }; h.conns[deviceID][c] = struct{}{}; h.mu.Unlock()
    h.Send(deviceID, "connected", map[string]any{"device_id":deviceID})
    defer func(){ h.mu.Lock(); delete(h.conns[deviceID],c); if len(h.conns[deviceID])==0 { delete(h.conns,deviceID) }; h.mu.Unlock(); _=c.Close(websocket.StatusNormalClosure,"closed") }()
    ctx:=r.Context()
    for { _,_,err:=c.Read(ctx); if err!=nil{return} }
}

func (h *RealtimeHub) Send(deviceID,event string,payload any){
    deviceID=strings.TrimSpace(deviceID); if deviceID=="" {return}
    data,err:=json.Marshal(map[string]any{"event":event,"payload":payload,"sent_at":time.Now().UTC()}); if err!=nil{return}
    h.mu.RLock(); conns:=make([]*websocket.Conn,0,len(h.conns[deviceID])); for c:=range h.conns[deviceID]{conns=append(conns,c)}; h.mu.RUnlock()
    for _,c:=range conns {ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);_ = c.Write(ctx,websocket.MessageText,data);cancel()}
}

var sessionMu sync.RWMutex
var sessions = map[string]string{}

func issueSession(deviceID string) string { id:=newSessionID(); if id==""{return ""}; sessionMu.Lock(); sessions[deviceID]=id; sessionMu.Unlock(); return id }
func validSession(deviceID,session string) bool { sessionMu.RLock(); defer sessionMu.RUnlock(); return session != "" && sessions[deviceID] == session }
func newSessionID() string { b:=make([]byte,18); if _,err:=rand.Read(b);err!=nil{return ""};return hex.EncodeToString(b) }
