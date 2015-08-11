package streamer

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// StreamID identifies a pair of standard output and error channels used for streaming.
type StreamID string

type Streamer interface {
	// Stream sets up streaming for the given pair of channels and returns a StreamID to identify the pair.
	// The caller must call Stop to avoid leaking memory.
	Stream(stdout, stderr chan []byte) StreamID

	// StreamStdout streams to the specified writer from the standard output channel of the specified pair of channels.
	StreamStdout(StreamID, io.Writer)

	// StreamStderr streams to the specified writer from the standard error channel of the specified pair of channels.
	StreamStderr(StreamID, io.Writer)

	// Stop stops streaming from the specified pair of channels.
	Stop(StreamID)
}

// New creates a Streamer with the specified grace time which limits the duration of memory consumption by a stopped stream.
func New(graceTime time.Duration) Streamer {
	return &streamer{
		graceTime: graceTime,
		streams:   make(map[StreamID]*stream),
	}
}

type streamer struct {
	mu           sync.Mutex
	nextStreamID uint64
	graceTime    time.Duration
	streams      map[StreamID]*stream
}

type stream struct {
	ch   [2]chan []byte
	done chan struct{}
}

func (m *streamer) Stream(stdout, stderr chan []byte) StreamID {
	m.mu.Lock()
	defer m.mu.Unlock()

	var sid StreamID = StreamID(fmt.Sprintf("%d", m.nextStreamID))
	m.nextStreamID++

	m.streams[sid] = &stream{
		ch:   [2]chan []byte{stdout, stderr},
		done: make(chan struct{}),
	}
	return sid
}

func (m *streamer) StreamStdout(streamID StreamID, writer io.Writer) {
	m.doStream(streamID, writer, 0)
}

func (m *streamer) StreamStderr(streamID StreamID, writer io.Writer) {
	m.doStream(streamID, writer, 1)
}

func (m *streamer) doStream(streamID StreamID, writer io.Writer, chanIndex int) {
	if writer == nil {
		return
	}
	strm := m.getStream(streamID)
	if strm == nil {
		return
	}

	streamUntilStopped(strm.ch[chanIndex], strm.done, writer)
}

func streamUntilStopped(ch chan []byte, done chan struct{}, writer io.Writer) {
	for {
		select {
		case b := <-ch:
			if _, err := writer.Write(b); err != nil {
				return
			}
		case <-done:
			drain(ch, writer)
			return
		}
	}
}

func drain(ch chan []byte, writer io.Writer) {
	drainCount := len(ch)
	for i := 0; i < drainCount; i++ {
		select {
		case b := <-ch:
			writer.Write(b)
		default:
		}
	}
}

func (m *streamer) Stop(streamID StreamID) {
	strm := m.getStream(streamID)
	if strm == nil {
		panic(fmt.Sprintf("Invalid stream ID %d", streamID))
	}
	close(strm.done)

	go func() {
		time.Sleep(m.graceTime)
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.streams, streamID)
	}()
}

func (m *streamer) getStream(streamID StreamID) *stream {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.streams[streamID]
}
