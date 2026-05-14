package server_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/valorisa/ShellFromBrowser/internal/server"
)

func TestWebSocketEcho(t *testing.T) {
	srv := server.New(":0")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer ws.Close()

	msg := []byte("hello")
	err = ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	_, got, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(msg) {
		t.Errorf("got %q, want %q", got, msg)
	}
}
