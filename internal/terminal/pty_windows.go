//go:build windows

package terminal

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
)

type winCloser struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (w *winCloser) Close() error {
	// Close stdin first to signal EOF to the process
	if w.stdin != nil {
		w.stdin.Close()
	}

	// Kill the process
	if w.cmd.Process != nil {
		w.cmd.Process.Kill()
	}

	// Wait for process to exit
	w.cmd.Wait()

	// stdout pipe will be closed automatically when process exits
	// Don't try to close it manually to avoid "file already closed" error
	return nil
}

func NewSession(cols, rows uint16) (*Session, error) {
	shell := os.Getenv("COMSPEC")
	if shell == "" {
		shell = "cmd.exe"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}

	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}

	id := make([]byte, 16)
	rand.Read(id)

	return &Session{
		id:     hex.EncodeToString(id),
		reader: stdout,
		writer: stdin,
		closer: &winCloser{cmd: cmd, stdin: stdin, stdout: stdout},
		cols:   cols,
		rows:   rows,
	}, nil
}
