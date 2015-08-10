package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type StreamServer struct {
	mu      sync.RWMutex
	nextID  uint64
	streams map[string]*s

	connectWait time.Duration
}

type s struct {
	stdout chan []byte
	stderr chan []byte
	done   chan struct{}
}

type stdoutOrErr bool

var (
	Stdout stdoutOrErr = true
	Stderr stdoutOrErr = false
)

func (t stdoutOrErr) pick(s *s) chan []byte {
	if t == Stdout {
		return s.stdout
	} else {
		return s.stderr
	}
}

func NewStreamServer(connectWait time.Duration) *StreamServer {
	return &StreamServer{
		streams:     make(map[string]*s),
		connectWait: connectWait,
	}
}

func (m *StreamServer) Stream(stdout, stderr chan []byte) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamID := fmt.Sprintf("%d", m.nextID)
	m.nextID++ // while this can technically overflow, if we created one process every single nanosecond, it would take approximately 600 years to do so

	m.streams[streamID] = &s{
		stdout: stdout,
		stderr: stderr,
		done:   make(chan struct{}),
	}

	return streamID
}

func (m *StreamServer) HandleStream(w http.ResponseWriter, r *http.Request, outOrErr stdoutOrErr) {
	streamid := r.FormValue(":streamid")

	w.WriteHeader(http.StatusOK)
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	m.mu.RLock()
	stream := m.streams[streamid]
	m.mu.RUnlock()

	streamAndDrain(conn, outOrErr.pick(stream), stream.done)
}

func (m *StreamServer) Stop(id string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stream := m.streams[id]
	close(stream.done)

	go func() {
		<-time.After(m.connectWait)

		// wait a little while to make sure the clients have started streaming, after which
		// the map key is unneeded (they've already done the lookup) so we can safely delete it
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.streams, id)
	}()
}

func streamAndDrain(conn io.Writer, ch chan []byte, done chan struct{}) {
	for {
		select {
		case output := <-ch:
			conn.Write(output)
		case <-done:
			for {
				select {
				case output := <-ch:
					conn.Write(output)
				default:
					return
				}
			}
		}
	}
}
