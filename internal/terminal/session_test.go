package terminal_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/valorisa/ShellFromBrowser/internal/terminal"
)

func TestSessionSpawnAndWrite(t *testing.T) {
	sess, err := terminal.NewSession(80, 24)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer sess.Close()

	// Write a command - use "echo hello" on Unix, or just send input on Windows
	_, err = sess.Write([]byte("echo hello\n"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Read output with timeout
	buf := make([]byte, 4096)
	var output bytes.Buffer
	deadline := time.After(5 * time.Second)

	for {
		select {
		case <-deadline:
			if !bytes.Contains(output.Bytes(), []byte("hello")) {
				t.Fatalf("timeout: output was %q", output.String())
			}
			return
		default:
			n, _ := sess.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
				if bytes.Contains(output.Bytes(), []byte("hello")) {
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestSessionClose(t *testing.T) {
	sess, err := terminal.NewSession(80, 24)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	err = sess.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Write after close should fail
	_, err = sess.Write([]byte("test"))
	if err == nil {
		t.Fatal("expected error writing to closed session")
	}
}
