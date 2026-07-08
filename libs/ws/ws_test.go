package ws

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// echoServer upgrades and echoes every text message back.
func echoServer(t *testing.T) (*httptest.Server, *sync.WaitGroup) {
	t.Helper()
	var wg sync.WaitGroup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Accept(w, r)
		if err != nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.Close()
			for {
				opcode, payload, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if opcode == OpText {
					if err := conn.WriteText(payload); err != nil {
						return
					}
				}
			}
		}()
	}))
	return server, &wg
}

func wsURL(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func TestDialEchoRoundTrip(t *testing.T) {
	server, wg := echoServer(t)
	defer server.Close()
	defer wg.Wait()

	conn, err := Dial(wsURL(server))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	messages := []string{
		"hello",
		strings.Repeat("m", 200),    // 16-bit extended length
		strings.Repeat("x", 70_000), // 64-bit extended length
	}
	for _, message := range messages {
		if err := conn.WriteText([]byte(message)); err != nil {
			t.Fatalf("WriteText: %v", err)
		}
		opcode, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage: %v", err)
		}
		if opcode != OpText || string(payload) != message {
			t.Fatalf("echo = opcode %v, %d bytes; want text %d bytes", opcode, len(payload), len(message))
		}
	}
}

func TestPingIsAnsweredTransparently(t *testing.T) {
	server, wg := echoServer(t)
	defer server.Close()
	defer wg.Wait()

	conn, err := Dial(wsURL(server))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// A ping from the client must be answered by the server's read loop
	// without disturbing the text stream.
	if err := conn.writeFrame(opPing, []byte("are-you-there")); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	if err := conn.WriteText([]byte("after-ping")); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	// The client read loop swallows the pong and returns the echoed text.
	opcode, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if opcode != OpText || string(payload) != "after-ping" {
		t.Fatalf("got %v %q", opcode, payload)
	}
}

func TestCloseHandshake(t *testing.T) {
	server, wg := echoServer(t)
	defer server.Close()
	defer wg.Wait()

	conn, err := Dial(wsURL(server))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, _, err := conn.ReadMessage(); !errors.Is(err, ErrClosed) {
		t.Fatalf("read after close err = %v, want ErrClosed", err)
	}
}

func TestAcceptRejectsPlainHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := Accept(w, r); !errors.Is(err, ErrNotWebSocket) {
			t.Errorf("Accept err = %v, want ErrNotWebSocket", err)
		}
	}))
	defer server.Close()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
}

func TestDialRejectsNonWebSocketServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if _, err := Dial(wsURL(server)); err == nil {
		t.Fatal("expected handshake failure")
	}
}
