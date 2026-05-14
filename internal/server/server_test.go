package server_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/valorisa/ShellFromBrowser/internal/config"
	"github.com/valorisa/ShellFromBrowser/internal/server"
)

func TestWebSocketEcho(t *testing.T) {
	srv := server.New(":0", config.Default())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer ws.Close()

	// Read initial banner/prompt from the PTY
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("read initial: %v", err)
	}

	// The websocket now spawns a real terminal session
	// Just verify we can connect and get output
	t.Log("WebSocket PTY connection established successfully")
}
