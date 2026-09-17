package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ExecutionResult struct {
	Action    string `json:"action"`
	ServiceID string `json:"service_id"`
	Unit      string `json:"unit"`
	Executed  bool   `json:"executed"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// executeRegistered performs only explicit systemd actions for a registered service.
// It never accepts a shell command from an API request.
func executeRegistered(ctx context.Context, svc Service, action string) ExecutionResult {
	result := ExecutionResult{Action: action, ServiceID: svc.ID, Unit: svc.Service, Executed: false}
	if strings.TrimSpace(svc.Service) == "" {
		result.Status = "rejected"
		result.Error = "registered systemd unit is empty"
		return result
	}
	if action != "deploy" && action != "rollback" && action != "restart" {
		result.Status = "rejected"
		result.Error = "action is not allowlisted"
		return result
	}

	var args []string
	switch action {
	case "deploy":
		args = []string{"restart", svc.Service}
	case "rollback":
		result.Status = "approval-recorded"
		result.Error = "rollback requires a registered artifact/version executor and is not implemented by systemd restart"
		return result
	case "restart":
		args = []string{"restart", svc.Service}
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("systemctl action failed: %v: %s", err, strings.TrimSpace(string(out)))
		return result
	}
	result.Executed = true
	result.Status = "executed"
	return result
}
