package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParseTarget(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "3000", want: "http://localhost:3000"},
		{in: "localhost:8080", want: "http://localhost:8080"},
		{in: "http://127.0.0.1:5000", want: "http://127.0.0.1:5000"},
		{in: "https://localhost:8443", want: "https://localhost:8443"},
		{in: "https://localhost", want: "https://localhost:443"},
		{in: "ftp://localhost:21", wantErr: true},
		{in: "99999", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, c := range cases {
		got, err := ParseTarget(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseTarget(%q) = %v, want error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseTarget(%q) returned %v", c.in, err)
			continue
		}
		if got.String() != c.want {
			t.Errorf("ParseTarget(%q) = %q, want %q", c.in, got.String(), c.want)
		}
	}
}

// fakeEdge speaks the edge side of the protocol so the client can be tested end to end.
type fakeEdge struct {
	t        *testing.T
	ln       net.Listener
	mu       sync.Mutex
	out      *conn
	reader   *bufio.Reader
	hello    helloMsg
	ready    chan struct{}
	lastAuth string
	bearer   string
	token    string
}

func newFakeEdge(t *testing.T, name string) *fakeEdge {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	e := &fakeEdge{t: t, ln: ln, ready: make(chan struct{})}
	e.hello = helloMsg{URL: "http://" + name + ".example.test", Token: "tok-" + name}
	go e.accept()
	t.Cleanup(func() { ln.Close() })
	return e
}

func (e *fakeEdge) url() string { return "http://" + e.ln.Addr().String() }

func (e *fakeEdge) accept() {
	c, err := e.ln.Accept()
	if err != nil {
		return
	}
	reader := bufio.NewReader(c)
	req, err := http.ReadRequest(reader)
	if err != nil {
		c.Close()
		return
	}
	e.mu.Lock()
	e.bearer = req.Header.Get("Authorization")
	e.lastAuth = req.Header.Get("X-Tunnel-Auth")
	e.token = req.Header.Get("X-Tunnel-Token")
	e.out = &conn{w: c}
	e.reader = reader
	e.mu.Unlock()

	fmt.Fprint(c, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: osir-tunnel\r\nConnection: Upgrade\r\n\r\n")
	e.out.sendJSON(frameHello, 0, e.hello)
	close(e.ready)
}

// roundTrip sends a raw request on a new stream and collects the response bytes.
func (e *fakeEdge) roundTrip(id uint32, request string) (string, error) {
	<-e.ready
	if err := e.out.sendJSON(frameOpen, id, openMsg{Remote: "198.51.100.7"}); err != nil {
		return "", err
	}
	if err := e.out.send(frameData, id, []byte(request)); err != nil {
		return "", err
	}

	var sb strings.Builder
	buf := make([]byte, maxPayload)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		f, err := readFrame(e.reader, buf)
		if err != nil {
			return sb.String(), err
		}
		switch f.typ {
		case frameData:
			if f.id == id {
				sb.Write(f.payload)
			}
		case frameEnd, frameClose:
			if f.id == id {
				return sb.String(), nil
			}
		case framePing:
			e.out.send(framePong, 0, nil)
		}
	}
	return sb.String(), fmt.Errorf("timed out waiting for the response")
}

// localServer is the app being exposed.
func localServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: handler}
	go srv.Serve(ln)
	t.Cleanup(func() { srv.Close() })
	return ln.Addr().String()
}

func runClient(t *testing.T, edge *fakeEdge, target string) (chan string, context.CancelFunc) {
	t.Helper()
	parsed, err := ParseTarget(target)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	urls := make(chan string, 4)
	go Run(ctx, Options{
		Server:   edge.url(),
		Target:   parsed,
		Auth:     "secret",
		Token:    func(context.Context) (string, error) { return "account-token", nil },
		OnReady:  func(u string) { urls <- u },
		OnStatus: func(msg string) { t.Log("status:", msg) },
	})
	t.Cleanup(cancel)
	return urls, cancel
}

func TestTunnelForwardsRequests(t *testing.T) {
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Host", r.Host)
		fmt.Fprintf(w, "hello from %s", r.URL.Path)
	})
	edge := newFakeEdge(t, "quiet-river")
	urls, _ := runClient(t, edge, addr)

	select {
	case got := <-urls:
		if got != edge.hello.URL {
			t.Fatalf("OnReady gave %q, want %q", got, edge.hello.URL)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("client never reported a URL")
	}

	resp, err := edge.roundTrip(1, "GET /demo HTTP/1.1\r\nHost: quiet-river.example.test\r\nConnection: close\r\n\r\n")
	if err != nil {
		t.Fatalf("round trip failed: %v (got %q)", err, resp)
	}
	if !strings.Contains(resp, "200 OK") || !strings.Contains(resp, "hello from /demo") {
		t.Fatalf("unexpected response: %q", resp)
	}
	if !strings.Contains(resp, "X-Seen-Host: quiet-river.example.test") {
		t.Fatalf("Host header did not reach the local server: %q", resp)
	}

	edge.mu.Lock()
	auth := edge.lastAuth
	edge.mu.Unlock()
	if auth != "secret" {
		t.Errorf("edge saw auth %q, want %q", auth, "secret")
	}
}

func TestTunnelCarriesLargeBodies(t *testing.T) {
	const size = 3 << 20 // 3 MiB
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprint(size))
		w.Write(make([]byte, size))
	})
	edge := newFakeEdge(t, "big-oak")
	runClient(t, edge, addr)

	resp, err := edge.roundTrip(1, "GET /big HTTP/1.1\r\nHost: big-oak.example.test\r\nConnection: close\r\n\r\n")
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
	body := resp[strings.Index(resp, "\r\n\r\n")+4:]
	if len(body) != size {
		t.Fatalf("got %d body bytes, want %d", len(body), size)
	}
}

func TestTunnelServes502WhenLocalServerIsDown(t *testing.T) {
	// Grab a port, then close it so nothing is listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dead := ln.Addr().String()
	ln.Close()

	edge := newFakeEdge(t, "lonely-pine")
	runClient(t, edge, dead)

	resp, err := edge.roundTrip(1, "GET / HTTP/1.1\r\nHost: lonely-pine.example.test\r\n\r\n")
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
	if !strings.Contains(resp, "502 Bad Gateway") || !strings.Contains(resp, "App not answering") {
		t.Fatalf("expected a 502 page, got %.400q", resp)
	}
	if !strings.Contains(resp, dead) {
		t.Errorf("502 page does not name the unreachable address %q", dead)
	}
}

func TestTunnelHandlesManyStreams(t *testing.T) {
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "path=%s", r.URL.Path)
	})
	edge := newFakeEdge(t, "many-streams")
	runClient(t, edge, addr)
	<-edge.ready

	// Streams are interleaved on purpose: open them all, then read the replies.
	const n = 25
	for i := 1; i <= n; i++ {
		if err := edge.out.sendJSON(frameOpen, uint32(i), openMsg{Remote: "198.51.100.7"}); err != nil {
			t.Fatal(err)
		}
		req := fmt.Sprintf("GET /p%d HTTP/1.1\r\nHost: many.example.test\r\nConnection: close\r\n\r\n", i)
		if err := edge.out.send(frameData, uint32(i), []byte(req)); err != nil {
			t.Fatal(err)
		}
	}

	seen := map[uint32]*strings.Builder{}
	finished := 0
	buf := make([]byte, maxPayload)
	deadline := time.Now().Add(10 * time.Second)
	for finished < n && time.Now().Before(deadline) {
		f, err := readFrame(edge.reader, buf)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		switch f.typ {
		case frameData:
			if seen[f.id] == nil {
				seen[f.id] = &strings.Builder{}
			}
			seen[f.id].Write(f.payload)
		case frameEnd:
			finished++
		case framePing:
			edge.out.send(framePong, 0, nil)
		}
	}
	if finished != n {
		t.Fatalf("only %d of %d streams finished", finished, n)
	}
	for i := uint32(1); i <= n; i++ {
		want := fmt.Sprintf("path=/p%d", i)
		if sb := seen[i]; sb == nil || !strings.Contains(sb.String(), want) {
			t.Errorf("stream %d did not carry %q", i, want)
		}
	}
}

func TestTunnelReconnectsWithItsToken(t *testing.T) {
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") })
	edge := newFakeEdge(t, "steady-fjord")
	runClient(t, edge, addr)
	<-edge.ready

	// Drop the connection; the client should come back presenting the token it was given.
	edge.mu.Lock()
	first := edge.out.w.(net.Conn)
	edge.mu.Unlock()
	edge.ready = make(chan struct{})
	go edge.accept()
	first.Close()

	select {
	case <-edge.ready:
	case <-time.After(10 * time.Second):
		t.Fatal("client did not reconnect")
	}
	edge.mu.Lock()
	token := edge.token
	edge.mu.Unlock()
	if token != "tok-steady-fjord" {
		t.Fatalf("reconnect presented token %q, want %q", token, "tok-steady-fjord")
	}
}

func TestHandshakeFailureIsFatal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		bufio.NewReader(c).ReadString('\n')
		fmt.Fprint(c, "HTTP/1.1 401 Unauthorized\r\nContent-Length: 18\r\nConnection: close\r\n\r\ninvalid auth token")
		c.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = Run(ctx, Options{Server: "http://" + ln.Addr().String(), Target: Target{Host: "localhost", Port: 3000}})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected a fatal 401 error, got %v", err)
	}
}

func TestHelloJSONRoundTrip(t *testing.T) {
	// The wire format must stay compatible with the Node edge.
	raw := `{"name":"quiet-river-7k2p","url":"https://quiet-river-7k2p.osir.run","token":"abc"}`
	var h helloMsg
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.URL != "https://quiet-river-7k2p.osir.run" || h.Token != "abc" {
		t.Fatalf("hello parsed wrong: %+v", h)
	}
}

func TestBadGatewayEscapesTarget(t *testing.T) {
	// The target comes from the command line, but it is echoed into a page shown in a
	// browser, so it must not be able to inject markup.
	page := string(badGateway(`http://x"><script>alert(1)</script>`))
	if strings.Contains(page, "<script>alert(1)</script>") {
		t.Fatalf("target was not escaped: %.300q", page)
	}
	if !strings.Contains(page, "&lt;script&gt;") {
		t.Fatalf("expected the escaped form in the page: %.300q", page)
	}
}

func TestTargetLoopbackDetection(t *testing.T) {
	cases := map[string]bool{
		"localhost": true, "LOCALHOST": true, "api.localhost": true,
		"127.0.0.1": true, "127.5.4.3": true, "::1": true,
		"example.com": false, "192.168.1.10": false, "0.0.0.0": false,
	}
	for host, want := range cases {
		got := Target{Host: host, Port: 443, UseTLS: true}.isLoopback()
		if got != want {
			t.Errorf("isLoopback(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestTunnelSendsTheAccountToken(t *testing.T) {
	addr := localServer(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") })
	edge := newFakeEdge(t, "signed-in")
	runClient(t, edge, addr)
	<-edge.ready

	edge.mu.Lock()
	bearer := edge.bearer
	edge.mu.Unlock()
	if bearer != "Bearer account-token" {
		t.Fatalf("edge saw Authorization %q, want %q", bearer, "Bearer account-token")
	}
}

func TestTunnelStopsWhenNotSignedIn(t *testing.T) {
	edge := newFakeEdge(t, "signed-out")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// An empty token must fail immediately rather than retry: no amount of reconnecting
	// turns a signed-out client into a signed-in one.
	err := Run(ctx, Options{
		Server: edge.url(),
		Target: Target{Host: "localhost", Port: 3000},
		Token:  func(context.Context) (string, error) { return "", nil },
	})
	if err == nil || !strings.Contains(err.Error(), "osir auth login") {
		t.Fatalf("expected a fatal not-signed-in error, got %v", err)
	}
}
