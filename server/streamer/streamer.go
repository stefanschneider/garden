package streamer

import (
	"fmt"
	"io"
	"math"
	"sync"
	"time"
)

// StreamID identifies a pair of standard output and error channels used for streaming.
type StreamID uint32

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
	return NewLimited(graceTime, math.MaxUint32)
}

// NewLimited creates a Streamer with the specified grace time (as described in New()) and maximum number of streams.
func NewLimited(graceTime time.Duration, maxStreams uint32) Streamer {
	return &streamer{
		graceTime:  graceTime,
		maxStreams: maxStreams,
		streams:    make(map[StreamID]*stream),
	}
}

type streamer struct {
	mu           sync.Mutex
	nextStreamID StreamID
	graceTime    time.Duration
	maxStreams   uint32
	streams      map[StreamID]*stream
}

type stream struct {
	stdout, stderr chan []byte
	done           chan struct{}
}

func (m *streamer) Stream(stdout, stderr chan []byte) StreamID {
	m.mu.Lock()
	defer m.mu.Unlock()
	found := false
	for i := uint32(0); !found && i < m.maxStreams; i++ {
		if m.streams[m.nextStreamID] != nil {
			m.getAndIncrementStreamID() // ignore return value
		} else {
			found = true
		}
	}

	if !found {
		panic("Number of streams cannot exceed the maximum allowed")
	}

	sid := m.getAndIncrementStreamID()
	m.streams[sid] = &stream{
		stdout: stdout,
		stderr: stderr,
		done:   make(chan struct{}),
	}
	return sid
}

func (m *streamer) getAndIncrementStreamID() StreamID {
	sid := m.nextStreamID
	m.nextStreamID++
	if uint32(m.nextStreamID) == m.maxStreams {
		m.nextStreamID = 0
	}
	return sid
}

func (m *streamer) StreamStdout(streamID StreamID, writer io.Writer) {
	if writer == nil {
		return
	}
	strm := m.getStream(streamID)
	if strm == nil {
		return
	}

	doStream(strm.stdout, strm.done, writer)
}

func doStream(ch chan []byte, done chan struct{}, writer io.Writer) {
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

func (m *streamer) remove(streamID StreamID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	strm := m.streams[streamID]
	if strm == nil {
		return
	}

	// Delete the stream if and only if it has been stopped.
	select {
	case <-strm.done:
		delete(m.streams, streamID)
	default:
	}
}

func (m *streamer) StreamStderr(streamID StreamID, writer io.Writer) {
	if writer == nil {
		return
	}
	strm := m.getStream(streamID)
	if strm == nil {
		return
	}

	doStream(strm.stderr, strm.done, writer)
}

func (m *streamer) Stop(streamID StreamID) {
	strm := m.getStream(streamID)
	if strm == nil {
		panic(fmt.Sprintf("Invalid stream ID %d", streamID))
	}
	close(strm.done)

	go func() {
		time.Sleep(m.graceTime)
		m.remove(streamID)
	}()
}

func (m *streamer) getStream(streamID StreamID) *stream {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.streams[streamID]
}
