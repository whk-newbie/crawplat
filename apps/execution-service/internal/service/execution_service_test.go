package service

import "testing"

func TestCreateManualExecutionStartsPending(t *testing.T) {
	svc := NewExecutionService()
	exec := svc.CreateManual("task-1", "spider-v1")
	if exec.Status != "pending" {
		t.Fatalf("expected pending, got %s", exec.Status)
	}
}

func TestCreateManualExecutionPersistsExecution(t *testing.T) {
	svc := NewExecutionService()
	created := svc.CreateManual("task-1", "spider-v1")

	listed, ok := svc.Get(created.ID)
	if !ok {
		t.Fatal("expected execution to be stored")
	}
	if listed.ID != created.ID || listed.TaskID != created.TaskID || listed.SpiderVersionID != created.SpiderVersionID || listed.Status != created.Status || listed.TriggerSource != created.TriggerSource {
		t.Fatalf("expected stored execution to match created execution, got %+v want %+v", listed, created)
	}
}

func TestAppendLogPersistsExecutionLog(t *testing.T) {
	svc := NewExecutionService()
	created := svc.CreateManual("task-1", "spider-v1")

	entry, err := svc.AppendLog(created.ID, "started")
	if err != nil {
		t.Fatalf("expected append log success, got error: %v", err)
	}
	if entry.Message != "started" {
		t.Fatalf("expected log message started, got %s", entry.Message)
	}

	logs, ok := svc.GetLogs(created.ID)
	if !ok {
		t.Fatal("expected logs to exist for execution")
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0] != entry {
		t.Fatalf("expected stored log to match appended log, got %+v want %+v", logs[0], entry)
	}
}
