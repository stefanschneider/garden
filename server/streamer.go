package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type StreamServer struct {
	mu      sync.RWMutex
	nextID  uint32
	streams map[uint32]*s

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
		streams:     make(map[uint32]*s),
		connectWait: connectWait,
	}
}

func (m *StreamServer) Stream(stdout, stderr chan []byte) uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	streamID := m.nextID
	m.nextID++

	m.streams[streamID] = &s{
		stdout: stdout,
		stderr: stderr,
		done:   make(chan struct{}),
	}

	return streamID
}

func (m *StreamServer) HandleStream(w http.ResponseWriter, r *http.Request, outOrErr stdoutOrErr) {
	streamid, err := strconv.Atoi(r.FormValue(":streamid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	m.mu.RLock()
	stream := m.streams[uint32(streamid)]
	ch := outOrErr.pick(stream)
	m.mu.RUnlock()

	for {
		select {
		case output := <-ch:
			conn.Write(output)
		case <-stream.done:
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

func (m *StreamServer) Stop(id uint32) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stream := m.streams[id]
	close(stream.done)

	go func() {
		<-time.After(m.connectWait) // allow clients time to join the wait group
		m.cleanup(id)               // delete stuff from the map now everyone is gone
	}()
}

func (m *StreamServer) cleanup(id uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.streams, id)
}
