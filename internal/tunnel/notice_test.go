package tunnel

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// The edge sends a notice before a planned restart, saying it is holding the address. That
// must reach the user as news and let the client reconnect, not end the session the way a
// refusal does.
func TestNoticeFromEdgeIsNotFatal(t *testing.T) {
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") })
	target, err := ParseTarget(addr)
	if err != nil {
		t.Fatal(err)
	}

	edge := newFakeEdge(t, "restarting-edge")
	notices := make(chan string, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go Run(ctx, Options{
		Server:   edge.url(),
		Target:   target,
		Token:    func(context.Context) (string, error) { return "account-token", nil },
		OnStatus: func(msg string) { notices <- msg },
	})
	<-edge.ready

	no := false
	if err := edge.out.sendJSON(frameError, 0, errorMsg{Message: "The edge is restarting; your address is held for 120s", Fatal: &no}); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(5 * time.Second)
	for {
		select {
		case msg := <-notices:
			if strings.Contains(msg, "restarting") {
				return // reported, and the client is still running
			}
		case <-deadline:
			t.Fatal("the edge's notice never reached the caller")
		}
	}
}

// A token error (say, the refresh could not reach Keycloak) is retried; only an empty token,
// meaning the session is really gone, stops the tunnel.
func TestTokenErrorIsRetriedUntilSignedOut(t *testing.T) {
	edge := newFakeEdge(t, "flaky-auth")
	calls := 0
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := Run(ctx, Options{
		Server: edge.url(),
		Target: Target{Host: "localhost", Port: 3000},
		Token: func(context.Context) (string, error) {
			if calls++; calls == 1 {
				return "", fmt.Errorf("cannot refresh your session right now: network is down")
			}
			return "", nil
		},
	})
	if calls != 2 {
		t.Fatalf("Token called %d times, want 2 (the error should have been retried)", calls)
	}
	if err == nil || !strings.Contains(err.Error(), "osir auth login") {
		t.Fatalf("expected a fatal not-signed-in error, got %v", err)
	}
}
