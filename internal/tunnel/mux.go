// Package tunnel implements the client side of the Osir Tunnel protocol: it keeps one
// connection to the tunnel edge and multiplexes every visitor connection over it, piping
// bytes verbatim to a local server.
//
// Frame layout: type(u8) | streamID(u32 BE) | length(u32 BE) | payload
package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Frame types. HELLO, OPEN and ERROR carry JSON payloads; the rest are control or raw data.
const (
	frameHello  = 1
	frameOpen   = 2
	frameData   = 3
	frameEnd    = 4
	frameClose  = 5
	framePause  = 6
	frameResume = 7
	framePing   = 8
	framePong   = 9
	frameError  = 10
)

const (
	headerLen  = 9
	maxPayload = 64 * 1024
	maxFrame   = 1024 * 1024
)

type frame struct {
	typ     byte
	id      uint32
	payload []byte
}

// helloMsg is sent by the edge once the tunnel is established. The edge also sends "name",
// which the URL already contains, so it is not decoded here.
type helloMsg struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// openMsg announces a new visitor connection.
type openMsg struct {
	Remote string `json:"remote"`
	TLS    bool   `json:"tls"`
}

// errorMsg is a message from the edge. Fatal is a pointer so that a missing field counts as
// fatal, while an explicit false marks a notice such as "restarting, your address is held".
type errorMsg struct {
	Message string `json:"message"`
	Fatal   *bool  `json:"fatal"`
}

// conn serializes frame writes onto the single edge connection.
type conn struct {
	w  io.Writer
	mu sync.Mutex
}

func (c *conn) send(typ byte, id uint32, payload []byte) error {
	if len(payload) > maxFrame {
		return fmt.Errorf("frame too large: %d bytes", len(payload))
	}
	buf := make([]byte, headerLen+len(payload))
	buf[0] = typ
	binary.BigEndian.PutUint32(buf[1:5], id)
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(payload)))
	copy(buf[headerLen:], payload)

	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.w.Write(buf)
	return err
}

func (c *conn) sendJSON(typ byte, id uint32, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.send(typ, id, payload)
}

// readFrame reads one frame. The payload is only valid until the next call.
func readFrame(r io.Reader, buf []byte) (frame, error) {
	var head [headerLen]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return frame{}, err
	}
	length := binary.BigEndian.Uint32(head[5:9])
	if length > maxFrame {
		return frame{}, fmt.Errorf("oversized frame: %d bytes", length)
	}
	f := frame{typ: head[0], id: binary.BigEndian.Uint32(head[1:5])}
	if length > 0 {
		if uint32(cap(buf)) < length {
			buf = make([]byte, length)
		}
		f.payload = buf[:length]
		if _, err := io.ReadFull(r, f.payload); err != nil {
			return frame{}, err
		}
	}
	return f, nil
}
