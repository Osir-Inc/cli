package tunnel

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	protocolVersion  = "1"
	pingEvery        = 20 * time.Second
	idleTimeout      = 60 * time.Second
	dialTimeout      = 15 * time.Second
	localDialTimeout = 10 * time.Second
	minBackoff       = time.Second
	maxBackoff       = 30 * time.Second
)

// Options configures a tunnel client.
type Options struct {
	Server string // tunnel edge, e.g. https://osir.run
	Target Target // local server to forward to, from ParseTarget
	Auth   string // shared secret for a private edge

	// Token returns the OSIR access token identifying the account opening the tunnel. It is
	// called for every connection attempt, so a refreshed token is picked up on reconnect.
	Token func(ctx context.Context) (string, error)

	// OnReady is called with the public URL once the tunnel is live. It fires again after a
	// reconnect only when the URL changed.
	OnReady func(url string)
	// OnStatus reports transient conditions (reconnects, local errors) for the user.
	OnStatus func(msg string)
}

// Target is a parsed local address to forward to.
type Target struct {
	Host   string
	Port   int
	UseTLS bool
}

func (t Target) String() string {
	scheme := "http"
	if t.UseTLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(t.Host, strconv.Itoa(t.Port)))
}

func (t Target) dial() (net.Conn, error) {
	addr := net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
	dialer := &net.Dialer{Timeout: localDialTimeout}
	if !t.UseTLS {
		return dialer.Dial("tcp", addr)
	}
	// Development servers on this machine usually have a self-signed certificate, so the
	// check is skipped for loopback only. A target anywhere else is verified normally.
	return tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName:         t.Host,
		InsecureSkipVerify: t.isLoopback(),
	})
}

// isLoopback reports whether the target is on this machine.
func (t Target) isLoopback() bool {
	if strings.EqualFold(t.Host, "localhost") || strings.HasSuffix(strings.ToLower(t.Host), ".localhost") {
		return true
	}
	ip := net.ParseIP(t.Host)
	return ip != nil && ip.IsLoopback()
}

// ParseTarget accepts a port, host:port, or a full http/https URL.
func ParseTarget(s string) (Target, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Target{}, fmt.Errorf("no local target given")
	}
	if port, err := strconv.Atoi(s); err == nil {
		if port < 1 || port > 65535 {
			return Target{}, fmt.Errorf("port out of range: %d", port)
		}
		return Target{Host: "localhost", Port: port}, nil
	}
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return Target{}, fmt.Errorf("cannot parse target %q: %w", s, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Target{}, fmt.Errorf("target must be http or https, got %q", u.Scheme)
	}
	t := Target{Host: u.Hostname(), UseTLS: u.Scheme == "https"}
	if t.Host == "" {
		return Target{}, fmt.Errorf("target has no host: %q", s)
	}
	if p := u.Port(); p != "" {
		t.Port, err = strconv.Atoi(p)
		if err != nil {
			return Target{}, fmt.Errorf("invalid port in %q", s)
		}
	} else if t.UseTLS {
		t.Port = 443
	} else {
		t.Port = 80
	}
	return t, nil
}

// fatalError marks failures that retrying cannot fix (bad token, wrong endpoint).
type fatalError struct{ msg string }

func (e *fatalError) Error() string { return e.msg }

// errAnnounced ends a session the edge warned about with a notice, so no second message is shown.
var errAnnounced = errors.New("edge closed after a notice")

// Run keeps a tunnel open until ctx is cancelled, reconnecting when the edge drops.
func Run(ctx context.Context, opts Options) error {
	target := opts.Target
	if target.Port == 0 {
		return fmt.Errorf("no local target given")
	}
	status := opts.OnStatus
	if status == nil {
		status = func(string) {}
	}

	var token, lastURL string
	backoff := minBackoff

	for {
		// Fires as soon as the edge names the tunnel, not when the session ends.
		onHello := func(url string) {
			backoff = minBackoff
			if url != lastURL {
				lastURL = url
				if opts.OnReady != nil {
					opts.OnReady(url)
				}
			}
		}

		err := session(ctx, opts, target, &token, status, onHello)

		if ctx.Err() != nil {
			return nil
		}
		var fatal *fatalError
		if errors.As(err, &fatal) {
			return fatal
		}
		if errors.Is(err, errAnnounced) {
			// The notice already told the user; a second "connection closed" line is noise.
		} else if err != nil {
			status(fmt.Sprintf("%v; reconnecting in %s", err, backoff.Round(time.Second)))
		} else {
			status(fmt.Sprintf("connection to the edge closed; reconnecting in %s", backoff.Round(time.Second)))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// session runs one connection to the edge and returns when it ends. onHello fires with the
// public URL as soon as the edge assigns one.
func session(ctx context.Context, opts Options, target Target, token *string, status func(string), onHello func(string)) error {
	netConn, reader, err := handshake(ctx, opts, *token)
	if err != nil {
		return err
	}
	defer netConn.Close()

	out := &conn{w: netConn}
	streams := make(map[uint32]*stream)
	var mu sync.Mutex

	closeAll := func() {
		mu.Lock()
		for _, s := range streams {
			s.close()
		}
		streams = map[uint32]*stream{}
		mu.Unlock()
	}
	defer closeAll()

	// Stop the read loop when the caller cancels.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			netConn.Close()
		case <-done:
		}
	}()

	// Keepalive: ping regularly. The read deadline below gives up if the edge goes quiet;
	// it lives in the read loop so a ping stuck on a dead socket can't hide the timeout.
	go func() {
		ticker := time.NewTicker(pingEvery)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := out.send(framePing, 0, nil); err != nil {
					netConn.Close()
					return
				}
			}
		}
	}()

	buf := make([]byte, maxPayload)
	noticed := false // the edge said why it is about to close
	for {
		netConn.SetReadDeadline(time.Now().Add(idleTimeout))
		f, err := readFrame(reader, buf)
		if err != nil {
			// The edge hung up, went quiet, or we were cancelled; the caller reconnects.
			// Returning closes netConn, which also unblocks any write stuck on it.
			if noticed {
				return errAnnounced
			}
			return nil
		}

		switch f.typ {
		case frameHello:
			var hello helloMsg
			if err := json.Unmarshal(f.payload, &hello); err != nil {
				return fmt.Errorf("bad hello from edge: %w", err)
			}
			*token = hello.Token
			onHello(hello.URL)

		case frameOpen:
			id := f.id
			s := newStream(id, out, target.dial, func(dialErr error) {
				status(fmt.Sprintf("cannot reach %s: %v", target, dialErr))
				out.send(frameData, id, badGateway(target.String()))
				out.send(frameEnd, id, nil)
			})
			mu.Lock()
			streams[id] = s
			mu.Unlock()

		case frameData:
			mu.Lock()
			s := streams[f.id]
			mu.Unlock()
			if s != nil && s.push(f.payload) {
				out.send(framePause, f.id, nil)
			}

		case frameEnd:
			mu.Lock()
			s := streams[f.id]
			mu.Unlock()
			if s != nil {
				s.end()
			}

		case frameClose:
			mu.Lock()
			s := streams[f.id]
			delete(streams, f.id)
			mu.Unlock()
			if s != nil {
				s.close()
			}

		case framePause, frameResume:
			mu.Lock()
			s := streams[f.id]
			mu.Unlock()
			if s != nil {
				s.setRemotePaused(f.typ == framePause)
			}

		case framePing:
			out.send(framePong, 0, nil)

		case framePong:
			// keepalive answered

		case frameError:
			var e errorMsg
			json.Unmarshal(f.payload, &e)
			if e.Fatal != nil && !*e.Fatal {
				// A notice: the edge is going away but will hold the address. Report it as
				// news rather than an error, and let the reconnect happen as usual.
				status(e.Message)
				noticed = true
				continue
			}
			return &fatalError{msg: "edge refused the tunnel: " + e.Message}
		}
	}
}

// handshake performs the HTTP upgrade and returns the raw connection plus a reader that
// holds any bytes already buffered after the 101 response.
func handshake(ctx context.Context, opts Options, token string) (net.Conn, *bufio.Reader, error) {
	u, err := url.Parse(opts.Server)
	if err != nil {
		return nil, nil, &fatalError{msg: fmt.Sprintf("invalid edge URL %q: %v", opts.Server, err)}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil, &fatalError{msg: fmt.Sprintf("edge URL must be http or https, got %q", u.Scheme)}
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	addr := net.JoinHostPort(u.Hostname(), port)

	dialer := &net.Dialer{Timeout: dialTimeout}
	var netConn net.Conn
	if u.Scheme == "https" {
		netConn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: u.Hostname(),
		})
	} else {
		netConn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("cannot reach %s (%v)", opts.Server, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(u.String(), "/")+"/_tunnel", nil)
	if err != nil {
		netConn.Close()
		return nil, nil, err
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "osir-tunnel")
	req.Header.Set("X-Tunnel-Version", protocolVersion)
	if opts.Token != nil {
		// An error is retried like any failed connection: the refresh may have failed only
		// because the network is down. The auth package clears a session that Keycloak
		// rejected, so the next attempt sees an empty token and stops.
		bearer, err := opts.Token(ctx)
		if err != nil {
			netConn.Close()
			return nil, nil, err
		}
		if bearer == "" {
			netConn.Close()
			return nil, nil, &fatalError{msg: "not signed in; run: osir auth login"}
		}
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if token != "" {
		req.Header.Set("X-Tunnel-Token", token)
	}
	if opts.Auth != "" {
		req.Header.Set("X-Tunnel-Auth", opts.Auth)
	}

	netConn.SetDeadline(time.Now().Add(dialTimeout))
	if err := req.Write(netConn); err != nil {
		netConn.Close()
		return nil, nil, fmt.Errorf("cannot start tunnel: %w", err)
	}
	reader := bufio.NewReader(netConn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		netConn.Close()
		return nil, nil, fmt.Errorf("no answer from %s: %w", opts.Server, err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		body := make([]byte, 512)
		n, _ := resp.Body.Read(body)
		netConn.Close()
		msg := fmt.Sprintf("edge refused the tunnel: %s %s", resp.Status, strings.TrimSpace(string(body[:n])))
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
			return nil, nil, &fatalError{msg: msg}
		}
		return nil, nil, fmt.Errorf("%s", msg)
	}
	netConn.SetDeadline(time.Time{})
	return netConn, reader, nil
}
