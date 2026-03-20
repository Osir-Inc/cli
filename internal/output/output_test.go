package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintSuccess_Text(t *testing.T) {
	var buf bytes.Buffer
	f := New(false)
	f.SetOut(&buf)

	f.PrintSuccess("done")

	if got := buf.String(); !strings.Contains(got, "[OK] done") {
		t.Errorf("PrintSuccess text = %q, want [OK] done", got)
	}
}

func TestPrintSuccess_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(true)
	f.SetOut(&buf)

	f.PrintSuccess("done")

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "success" {
		t.Errorf("status = %q, want success", result["status"])
	}
	if result["message"] != "done" {
		t.Errorf("message = %q, want done", result["message"])
	}
}

func TestPrintError_Text(t *testing.T) {
	var buf bytes.Buffer
	f := New(false)
	f.SetErr(&buf)

	f.PrintError("failed")

	if got := buf.String(); !strings.Contains(got, "[ERROR] failed") {
		t.Errorf("PrintError text = %q, want [ERROR] failed", got)
	}
}

func TestPrintError_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(true)
	f.SetErr(&buf)

	f.PrintError("failed")

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "error" {
		t.Errorf("status = %q, want error", result["status"])
	}
}

func TestPrintTable_Text(t *testing.T) {
	var buf bytes.Buffer
	f := New(false)
	f.SetOut(&buf)

	headers := []string{"ID", "NAME"}
	rows := [][]string{{"1", "example.com"}, {"2", "test.net"}}
	f.PrintTable(headers, rows)

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Error("table should contain headers")
	}
	if !strings.Contains(out, "example.com") || !strings.Contains(out, "test.net") {
		t.Error("table should contain row data")
	}
	if !strings.Contains(out, "---") {
		t.Error("table should contain separator line")
	}
}

func TestPrintTable_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(true)
	f.SetOut(&buf)

	headers := []string{"A", "B"}
	rows := [][]string{{"1", "2"}}
	f.PrintTable(headers, rows)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["headers"] == nil || result["rows"] == nil {
		t.Error("JSON table should have headers and rows keys")
	}
}

func TestPrintKeyValue_Text(t *testing.T) {
	var buf bytes.Buffer
	f := New(false)
	f.SetOut(&buf)

	f.PrintKeyValue("Domain", "example.com")

	out := buf.String()
	if !strings.Contains(out, "Domain:") || !strings.Contains(out, "example.com") {
		t.Errorf("PrintKeyValue = %q, want Domain: example.com", out)
	}
}

func TestIsJSON(t *testing.T) {
	f := New(true)
	if !f.IsJSON() {
		t.Error("expected IsJSON() = true")
	}
	f2 := New(false)
	if f2.IsJSON() {
		t.Error("expected IsJSON() = false")
	}
}

func TestPrintResult_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(true)
	f.SetOut(&buf)

	data := map[string]string{"key": "value"}
	f.PrintResult(data)

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("key = %q, want value", result["key"])
	}
}
