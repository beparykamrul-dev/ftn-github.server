package main

import (
	"context"
	"testing"
)

func TestExecutorRejectsUnknownAction(t *testing.T) {
	r := executeRegistered(context.Background(), Service{ID: "svc-1", Service: "ftn-test.service"}, "shell")
	if r.Executed || r.Status != "rejected" { t.Fatalf("expected rejection, got %+v", r) }
}

func TestExecutorDoesNotImplementRollbackAsRestart(t *testing.T) {
	r := executeRegistered(context.Background(), Service{ID: "svc-1", Service: "ftn-test.service"}, "rollback")
	if r.Executed || r.Status != "approval-recorded" { t.Fatalf("expected rollback to remain disabled, got %+v", r) }
}
