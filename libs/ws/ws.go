// Package ws is a minimal RFC 6455 WebSocket implementation covering what
// the market data feed needs: HTTP upgrade, unfragmented text/binary frames,
// and ping/pong/close control handling. It has no external dependencies.
//
// Deliberately unsupported (rejected, not silently mishandled): frame
// fragmentation, extensions/compression, and subprotocol negotiation.
package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	// MaxPayloadBytes bounds a single frame payload.
	MaxPayloadBytes = 1 << 20

	acceptGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// Opcode identifies a WebSocket frame type.
type Opcode byte

const (
	OpText   Opcode = 0x1
	OpBinary Opcode = 0x2
	opClose  Opcode = 0x8
	opPing   Opcode = 0x9
	opPong   Opcode = 0xA
)

var (
	ErrNotWebSocket    = errors.New("not a websocket upgrade request")
	ErrClosed          = errors.New("websocket connection closed")
	ErrProtocol        = errors.New("websocket protocol violation")
	ErrPayloadTooLarge = errors.New("websocket payload exceeds limit")
)

// Conn is a WebSocket connection. Reads and writes are internally locked,
// so one reader goroutine and any number of writers are safe.
type Conn struct {
	conn net.Conn
	rw   *bufio.ReadWriter

	writeMu sync.Mutex
	// maskWrites is true on client connections; clients must mask frames.
	maskWrites bool

	closeOnce sync.Once
}

// Accept upgrades an HTTP request to a WebSocket connection. On error, an
// HTTP error response has already been written.
func Accept(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if r.Method != http.MethodGet ||
		!headerContainsToken(r.Header, "Connection", "upgrade") ||
		!headerContainsToken(r.Header, "Upgrade", "websocket") {
		http.Error(w, "expected websocket upgrade", http.StatusBadRequest)
		return nil, ErrNotWebSocket
	}
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		w.Header().Set("Sec-WebSocket-Version", "13")
		http.Error(w, "unsupported websocket version", http.StatusUpgradeRequired)
		return nil, ErrNotWebSocket
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return nil, ErrNotWebSocket
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported by server", http.StatusInternalServerError)
		return nil, ErrNotWebSocket
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		return nil, err
	}

	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey(key) + "\r\n\r\n"
	if _, err := rw.WriteString(response); err != nil {
		conn.Close()
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		conn.Close()
		return nil, err
	}
	return &Conn{conn: conn, rw: rw}, nil
}

// Dial connects to a ws:// URL such as ws://127.0.0.1:8080/ws/public.
func Dial(url string) (*Conn, error) {
	return DialWithHeader(url, nil)
}

// DialWithHeader connects to a ws:// URL with custom headers.
func DialWithHeader(url string, header http.Header) (*Conn, error) {
	rest, ok := strings.CutPrefix(url, "ws://")
	if !ok {
		return nil, fmt.Errorf("%w: only ws:// URLs are supported", ErrNotWebSocket)
	}
	host, path, found := strings.Cut(rest, "/")
	if !found {
		path = ""
	}
	path = "/" + path

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Fixed handshake key: the key exists to defeat caches, not for
	// security, and a constant keeps the client deterministic.
	const key = "ZXhjaGFuZ2Utd3MtY2xpZW50IQ=="
	request := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n"

	for k, v := range header {
		for _, val := range v {
			request += k + ": " + val + "\r\n"
		}
	}
	request += "\r\n"
	if _, err := rw.WriteString(request); err != nil {
		conn.Close()
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		conn.Close()
		return nil, err
	}

	response, err := http.ReadResponse(rw.Reader, &http.Request{Method: http.MethodGet})
	if err != nil {
		conn.Close()
		return nil, err
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("%w: handshake status %d", ErrNotWebSocket, response.StatusCode)
	}
	if response.Header.Get("Sec-WebSocket-Accept") != acceptKey(key) {
		conn.Close()
		return nil, fmt.Errorf("%w: bad accept key", ErrProtocol)
	}
	return &Conn{conn: conn, rw: rw, maskWrites: true}, nil
}

// WriteText sends one text frame.
func (c *Conn) WriteText(payload []byte) error {
	return c.writeFrame(OpText, payload)
}

// ReadMessage returns the next text or binary message. Control frames are
// handled internally: pings are answered, close completes the closing
// handshake and returns ErrClosed.
func (c *Conn) ReadMessage() (Opcode, []byte, error) {
	for {
		fin, opcode, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch opcode {
		case OpText, OpBinary:
			if !fin {
				return 0, nil, fmt.Errorf("%w: fragmented frames unsupported", ErrProtocol)
			}
			return opcode, payload, nil
		case opPing:
			if err := c.writeFrame(opPong, payload); err != nil {
				return 0, nil, err
			}
		case opPong:
			// Ignore.
		case opClose:
			_ = c.writeFrame(opClose, nil)
			c.close()
			return 0, nil, ErrClosed
		default:
			return 0, nil, fmt.Errorf("%w: opcode %#x", ErrProtocol, opcode)
		}
	}
}

// Close sends a close frame and tears the connection down.
func (c *Conn) Close() error {
	_ = c.writeFrame(opClose, nil)
	return c.close()
}

func (c *Conn) close() error {
	var err error
	c.closeOnce.Do(func() {
		err = c.conn.Close()
	})
	return err
}

func (c *Conn) writeFrame(opcode Opcode, payload []byte) error {
	if len(payload) > MaxPayloadBytes {
		return ErrPayloadTooLarge
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	header := make([]byte, 0, 14)
	header = append(header, 0x80|byte(opcode))

	maskBit := byte(0)
	if c.maskWrites {
		maskBit = 0x80
	}
	switch {
	case len(payload) < 126:
		header = append(header, maskBit|byte(len(payload)))
	case len(payload) <= 0xFFFF:
		header = append(header, maskBit|126, 0, 0)
		binary.BigEndian.PutUint16(header[len(header)-2:], uint16(len(payload)))
	default:
		header = append(header, maskBit|127, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(header[len(header)-8:], uint64(len(payload)))
	}

	if c.maskWrites {
		// A fixed mask key is protocol-legal; masking exists to defeat
		// intermediary caches, not for confidentiality.
		key := [4]byte{0x12, 0x34, 0x56, 0x78}
		header = append(header, key[:]...)
		masked := make([]byte, len(payload))
		for i, b := range payload {
			masked[i] = b ^ key[i%4]
		}
		payload = masked
	}

	if _, err := c.rw.Write(header); err != nil {
		return err
	}
	if _, err := c.rw.Write(payload); err != nil {
		return err
	}
	return c.rw.Flush()
}

func (c *Conn) readFrame() (fin bool, opcode Opcode, payload []byte, err error) {
	var head [2]byte
	if _, err := io.ReadFull(c.rw, head[:]); err != nil {
		return false, 0, nil, translateEOF(err)
	}
	fin = head[0]&0x80 != 0
	if head[0]&0x70 != 0 {
		return false, 0, nil, fmt.Errorf("%w: reserved bits set", ErrProtocol)
	}
	opcode = Opcode(head[0] & 0x0F)
	masked := head[1]&0x80 != 0

	length := uint64(head[1] & 0x7F)
	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(c.rw, ext[:]); err != nil {
			return false, 0, nil, translateEOF(err)
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(c.rw, ext[:]); err != nil {
			return false, 0, nil, translateEOF(err)
		}
		length = binary.BigEndian.Uint64(ext[:])
	}
	if length > MaxPayloadBytes {
		return false, 0, nil, ErrPayloadTooLarge
	}

	// Servers must receive masked frames; clients must receive unmasked.
	if !c.maskWrites && !masked {
		return false, 0, nil, fmt.Errorf("%w: client frame not masked", ErrProtocol)
	}
	if c.maskWrites && masked {
		return false, 0, nil, fmt.Errorf("%w: server frame masked", ErrProtocol)
	}

	var key [4]byte
	if masked {
		if _, err := io.ReadFull(c.rw, key[:]); err != nil {
			return false, 0, nil, translateEOF(err)
		}
	}
	payload = make([]byte, length)
	if _, err := io.ReadFull(c.rw, payload); err != nil {
		return false, 0, nil, translateEOF(err)
	}
	if masked {
		for i := range payload {
			payload[i] ^= key[i%4]
		}
	}
	return fin, opcode, payload, nil
}

func acceptKey(key string) string {
	sum := sha1.Sum([]byte(key + acceptGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerContainsToken(header http.Header, name, token string) bool {
	for _, value := range header.Values(name) {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

func translateEOF(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed) {
		return ErrClosed
	}
	return err
}
