package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/osir/cli/internal/config"
)

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("expected /test, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	cfg := &config.Config{BackendURL: server.URL}
	c := NewClient(cfg, nil)

	var result map[string]string
	err := c.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
}

func TestGet_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("expected page=1, got %s", r.URL.Query().Get("page"))
		}
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	cfg := &config.Config{BackendURL: server.URL}
	c := NewClient(cfg, nil)

	q := make(map[string][]string)
	q["page"] = []string{"1"}
	var result map[string]any
	err := c.Get(context.Background(), "/test", q, &result)
	if err != nil {
		t.Fatalf("Get with query failed: %v", err)
	}
}

func TestPost_SendsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("body.name = %q, want test", body["name"])
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	cfg := &config.Config{BackendURL: server.URL}
	c := NewClient(cfg, nil)

	var result map[string]bool
	err := c.Post(context.Background(), "/create", map[string]string{"name": "test"}, &result)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if !result["success"] {
		t.Error("expected success=true")
	}
}

func TestGet_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	cfg := &config.Config{BackendURL: server.URL}
	c := NewClient(cfg, nil)

	var result map[string]any
	err := c.Get(context.Background(), "/missing", nil, &result)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestDelete_Method(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	cfg := &config.Config{BackendURL: server.URL}
	c := NewClient(cfg, nil)

	var result map[string]bool
	err := c.Delete(context.Background(), "/resource", &result)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
