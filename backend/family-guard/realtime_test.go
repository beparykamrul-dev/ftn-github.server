package main

import (
    "testing"
)

func TestSessionIssueAndValidation(t *testing.T) {
    deviceID := "android-test-realtime"
    session := issueSession(deviceID)
    if session == "" {
        t.Fatal("expected session")
    }
    if !validSession(deviceID, session) {
        t.Fatal("expected issued session to validate")
    }
    if validSession(deviceID, "wrong-session") {
        t.Fatal("wrong session must not validate")
    }
    if validSession("other-device", session) {
        t.Fatal("session must be bound to device")
    }
}

func TestRealtimeHubCanBeCreated(t *testing.T) {
    h := NewRealtimeHub()
    if h == nil || h.conns == nil {
        t.Fatal("expected initialized realtime hub")
    }
}
