package tunnel

import (
	"io"
	"net"
	"sync"
)

// Flow control: when a local server can't keep up, queued bytes build up here. Past the
// high-water mark we ask the edge to stop reading from the visitor, and resume once drained.
const (
	pauseAbove  = 1 << 20 // 1 MiB queued for the local server
	resumeBelow = 1 << 18 // 256 KiB
)

// stream carries one visitor connection: bytes from the edge are queued and written to the
// local server, and everything the local server sends goes back as DATA frames.
type stream struct {
	id    uint32
	out   *conn
	local net.Conn      // set once the local server answers
	ready chan struct{} // closed when local is set, or when the dial failed

	mu          sync.Mutex
	cond        *sync.Cond
	queue       [][]byte
	queued      int
	ended       bool // edge signalled end of the visitor's request body
	closed      bool
	weSentPause bool

	remotePaused bool // the edge asked us to stop sending
	remoteCond   *sync.Cond
}

// newStream registers a visitor connection right away and dials the local server in the
// background, so frames that arrive while dialing are queued rather than lost.
func newStream(id uint32, out *conn, dial func() (net.Conn, error), onDialError func(error)) *stream {
	s := &stream{id: id, out: out, ready: make(chan struct{})}
	s.cond = sync.NewCond(&s.mu)
	s.remoteCond = sync.NewCond(&s.mu)

	go func() {
		local, err := dial()
		if err != nil {
			onDialError(err)
			s.mu.Lock()
			s.closed = true
			s.cond.Broadcast()
			s.remoteCond.Broadcast()
			s.mu.Unlock()
			close(s.ready)
			return
		}
		s.mu.Lock()
		s.local = local
		aborted := s.closed
		s.mu.Unlock()
		close(s.ready)
		if aborted {
			local.Close()
			return
		}
		go s.writeToLocal()
		go s.readFromLocal()
	}()
	return s
}

// push queues bytes for the local server. Returns true if the edge should pause.
func (s *stream) push(data []byte) bool {
	chunk := make([]byte, len(data))
	copy(chunk, data)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.queue = append(s.queue, chunk)
	s.queued += len(chunk)
	s.cond.Signal()
	if s.queued > pauseAbove && !s.weSentPause {
		s.weSentPause = true
		return true
	}
	return false
}

func (s *stream) end() {
	s.mu.Lock()
	s.ended = true
	s.cond.Signal()
	s.mu.Unlock()
}

func (s *stream) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	local := s.local
	s.cond.Broadcast()
	s.remoteCond.Broadcast()
	s.mu.Unlock()
	if local != nil {
		local.Close()
	}
}

func (s *stream) setRemotePaused(paused bool) {
	s.mu.Lock()
	s.remotePaused = paused
	s.remoteCond.Broadcast()
	s.mu.Unlock()
}

// writeToLocal drains the queue into the local server, one chunk at a time.
func (s *stream) writeToLocal() {
	for {
		s.mu.Lock()
		for len(s.queue) == 0 && !s.ended && !s.closed {
			s.cond.Wait()
		}
		if s.closed {
			s.mu.Unlock()
			return
		}
		if len(s.queue) == 0 { // ended and drained: half-close so the app sees EOF
			s.mu.Unlock()
			if cw, ok := s.local.(interface{ CloseWrite() error }); ok {
				cw.CloseWrite()
			}
			return
		}
		chunk := s.queue[0]
		s.queue = s.queue[1:]
		s.queued -= len(chunk)
		resume := s.weSentPause && s.queued < resumeBelow
		if resume {
			s.weSentPause = false
		}
		s.mu.Unlock()

		if _, err := s.local.Write(chunk); err != nil {
			s.out.send(frameClose, s.id, nil)
			s.close()
			return
		}
		if resume {
			s.out.send(frameResume, s.id, nil)
		}
	}
}

// readFromLocal forwards the local server's response back to the edge.
func (s *stream) readFromLocal() {
	buf := make([]byte, maxPayload)
	for {
		s.mu.Lock()
		for s.remotePaused && !s.closed {
			s.remoteCond.Wait()
		}
		closed := s.closed
		s.mu.Unlock()
		if closed {
			return
		}

		n, err := s.local.Read(buf)
		if n > 0 {
			if sendErr := s.out.send(frameData, s.id, buf[:n]); sendErr != nil {
				s.close()
				return
			}
		}
		if err != nil {
			if err == io.EOF {
				s.out.send(frameEnd, s.id, nil)
			} else {
				s.out.send(frameClose, s.id, nil)
				s.close()
			}
			return
		}
	}
}
