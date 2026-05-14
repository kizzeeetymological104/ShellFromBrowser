package terminal

import (
	"io"
	"sync"
	"time"
)

type Session struct {
	id      string
	reader  io.Reader
	writer  io.Writer
	closer  io.Closer
	cols    uint16
	rows    uint16
	mu      sync.Mutex
	closed  bool
	resizer Resizer
}

type Resizer interface {
	Resize(cols, rows uint16) error
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	return s.writer.Write(p)
}

func (s *Session) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *Session) SetReadDeadline(t time.Time) {
	// No-op for pipe-based implementations
}

func (s *Session) Resize(cols, rows uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cols = cols
	s.rows = rows
	if s.resizer != nil {
		return s.resizer.Resize(cols, rows)
	}
	return nil
}

func (s *Session) Cols() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cols
}

func (s *Session) Rows() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rows
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.closer.Close()
}
