package streamer_test

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	"runtime"

	"github.com/cloudfoundry-incubator/garden/server/streamer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Streamer", func() {

	const testString = "x"

	var (
		graceTime time.Duration
		str       *streamer.Streamer

		stdoutChan        chan []byte
		stderrChan        chan []byte
		testByteSlice     []byte
		channelBufferSize int
	)

	JustBeforeEach(func() {
		str = streamer.New(graceTime)
		stdoutChan = make(chan []byte, channelBufferSize)
		stderrChan = make(chan []byte, channelBufferSize)
	})

	BeforeEach(func() {
		graceTime = 10 * time.Second
		channelBufferSize = 1

		testByteSlice = []byte(testString)
	})

	It("should stream standard output until it is stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
		}
		go str.ServeStdout(sid, w)
		stdoutChan <- testByteSlice
		stdoutChan <- testByteSlice
		Eventually(w.String).Should(Equal("xx"))
		str.Stop(sid)
	})

	It("should stream standard error until it is stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
		}
		go str.ServeStderr(sid, w)
		stderrChan <- testByteSlice
		stderrChan <- testByteSlice
		Eventually(w.String).Should(Equal("xx"))
		str.Stop(sid)
	})

	Describe("draining", func() {
		var (
			drain       streamer.DrainFunc
			drainCalled chan struct{}
		)

		JustBeforeEach(func() {
			drainCalled = make(chan struct{})
			drain = func(ch chan []byte, writer io.Writer) {
				close(drainCalled)
			}

			str = streamer.NewWithDrain(graceTime, drain)
		})

		It("should call the drain function for standard output after being stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			w := &syncBuffer{
				Buffer: new(bytes.Buffer),
			}
			str.ServeStdout(sid, w)
			Eventually(drainCalled).Should(BeClosed())
		})

		It("should call the drain function for standard error after being stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			w := &syncBuffer{
				Buffer: new(bytes.Buffer),
			}
			str.ServeStderr(sid, w)
			Eventually(drainCalled).Should(BeClosed())
		})
	})

	Context("when a grace time has been set", func() {
		BeforeEach(func() {
			graceTime = 100 * time.Millisecond

		})

		It("should not leak unused streams for longer than the grace time after streaming has been stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			time.Sleep(200 * time.Millisecond)
			Expect(func() { str.Stop(sid) }).To(Panic(), "stream was not removed")
		})

		It("should not leak goroutines for longer than the grace time after streaming has been stopped", func() {
			initialNumGoroutine := runtime.NumGoroutine()

			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)

			Eventually(func() int {
				return runtime.NumGoroutine()
			}, "200ms").Should(Equal(initialNumGoroutine))
		})
	})

	It("should terminate streaming output after a write error has occurred", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
			fail:   true,
		}
		go str.ServeStdout(sid, w)
		stdoutChan <- testByteSlice
		stdoutChan <- testByteSlice
		Consistently(w.String).Should(Equal(""))
		str.Stop(sid)
	})

	It("should terminate streaming errors after a write error has occurred", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
			fail:   true,
		}
		go str.ServeStderr(sid, w)
		stderrChan <- testByteSlice
		stderrChan <- testByteSlice
		Consistently(w.String).Should(Equal(""))
		str.Stop(sid)
	})
})

type syncBuffer struct {
	*bytes.Buffer
	fail bool
	mu   sync.Mutex
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if sb.fail {
		sb.fail = false
		return 0, errors.New("failed")
	}
	return sb.Buffer.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.Buffer.String()
}
